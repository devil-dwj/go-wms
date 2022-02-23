package pms

import (
	"fmt"

	"github.com/devil-dwj/go-wms/cmd/protoc-gen-routers/generator"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"google.golang.org/protobuf/proto"

	// pb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	options "google.golang.org/genproto/googleapis/api/annotations"
)

func init() {
	generator.RegisterPlugin(new(pms))
}

// 生成路由文件的插件实现
type pms struct {
	gen *generator.Generator
}

// 返回此插件名称
func (g *pms) Name() string {
	return "jjpms"
}

// 插件初始化 给generator调用
func (g *pms) Init(gen *generator.Generator) {
	g.gen = gen
}

// 给定定义的类型名，返回目标对象
func (g *pms) objectNamed(name string) generator.Object {
	// g.gen.RecordTypeUse(name) // 记录使用
	return g.gen.ObjectNamed(name)
}

// 给定定义的.proto名称，返回类型名
func (g *pms) typeName(str string) string {
	return g.gen.TypeName(g.objectNamed(str))
}

// P 转发给 g.gen.P.
func (g *pms) P(args ...interface{}) { g.gen.P(args...) }

// 生成给定文件中的service代码
func (g *pms) Generate(file *generator.FileDescriptor) {

	for i, service := range file.FileDescriptorProto.Service {
		g.generateService(file, service, i)
	}
}

// GenerateImports 生成文件导入
func (g *pms) GenerateImports(file *generator.FileDescriptor, imports map[generator.GoImportPath]generator.GoPackageName) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
}

// generateService 为service生成所有代码
func (g *pms) generateService(file *generator.FileDescriptor, service *descriptor.ServiceDescriptorProto, index int) {
	origServName := service.GetName()

	// 转成驼峰命名
	// eg.
	// 1. user_service => UserService
	// 2. user => User
	// 3. User => User
	servName := generator.CamelCase(origServName)
	g.P()

	// handler interface
	g.P(fmt.Sprintf("type %sHandler interface {", servName))

	for _, method := range service.Method {
		meth, path := g.getRule(method)
		if meth == "" || path == "" {
			continue
		}

		var (
			argInput = ""
			argOut   = ""
		)

		inputTypeName := g.typeName(method.GetInputType())
		outTypeName := g.typeName(method.GetOutputType())

		if !typeEmpty(inputTypeName) {
			argInput = fmt.Sprintf("req *%s", inputTypeName)
		}

		if !typeEmpty(outTypeName) {
			argOut = fmt.Sprintf("*%s, ", outTypeName)
		}

		g.P(fmt.Sprintf("%s(%s) (%serror)",
			method.GetName(), argInput, argOut))

	}
	g.P("}")
	g.P()

	// router
	g.P(fmt.Sprintf("type %sRouter interface {", servName))
	for _, method := range service.Method {
		g.P(fmt.Sprintf("%s(c *gin.Context)", method.GetName()))
	}
	g.P("}")
	g.P()

	g.P(fmt.Sprintf("type %s_Router struct {", servName))
	g.P(servName, "Handler")
	g.P("}")

	// register routers
	g.P(fmt.Sprintf("func Register%sRouters(a api.Api, h %sHandler) {", servName, servName))
	g.P(fmt.Sprintf("r := &%s_Router{h}", servName))

	for _, method := range service.Method {
		if method.Options != nil && proto.HasExtension(method.Options, options.E_Http) {
			meth, path := g.getRule(method)
			g.P(fmt.Sprintf("a.%s(\"%s\", r.%s)", meth, path, method.GetName()))
		}
	}

	g.P("}")
	g.P()

	// impl routers
	for _, method := range service.Method {
		g.P(fmt.Sprintf("func (h *%s_Router) %s(c *gin.Context) {", servName, method.GetName()))

		meth, path := g.getRule(method)
		if meth == "" || path == "" {
			continue
		}

		if meth == "POST" {
			// post

			var (
				argReq = ""
			)

			inputTypeName := g.typeName(method.GetInputType())
			outTypeName := g.typeName(method.GetOutputType())

			if !typeEmpty(inputTypeName) {
				argReq = "req"

				g.P(fmt.Sprintf("req := &%s{}", g.typeName(method.GetInputType())))
				g.P("if err := c.ShouldBindJSON(req); err != nil {")
				g.P("h.fail(c, err)")
				g.P("return")
				g.P("}")
				g.P()
			}

			if typeEmpty(outTypeName) {
				g.P(fmt.Sprintf("err := h.%sHandler.%s(%s)", servName, method.GetName(), argReq))
				g.P("if err != nil {")
				g.P("h.fail(c, err)")
				g.P("return")
				g.P("}")
				g.P()
				g.P("h.returnBack(c, err, \"\")")
				g.P("}")
				g.P()

			} else {
				g.P(fmt.Sprintf("rsp, err := h.%sHandler.%s(%s)", servName, method.GetName(), argReq))
				g.P("if err != nil {")
				g.P("h.fail(c, err)")
				g.P("return")
				g.P("}")
				g.P()
				g.P("h.returnBack(c, err, rsp)")
				g.P("}")
				g.P()
			}
		} else if meth == "GET" {
			// GET
			inputTypeName := g.typeName(method.GetInputType())
			if !typeEmpty(inputTypeName) {
				// get 请求有参数
				mProto := g.gen.GetDescriptorProto(inputTypeName)
				if mProto != nil {

					arg := "\n"
					// ide := ", "
					reqField := ""

					for _, field := range mProto.GetField() {
						fieldName := field.GetName()

						if field.GetType() == descriptor.FieldDescriptorProto_TYPE_INT32 {
							g.P(fmt.Sprintf(`%s, err := strconv.Atoi(c.Query("%s"))`, fieldName, fieldName))
							g.P("if err != nil {")
							g.P("h.fail(c, err)")
							g.P("return")
							g.P("}")
							g.P()

							reqField = fmt.Sprintf("%s: int32(%s), \n", generator.CamelCase(fieldName), fieldName)

						} else if field.GetType() == descriptor.FieldDescriptorProto_TYPE_STRING {
							g.P(fmt.Sprintf(`%s := c.Query("%s")`, fieldName, fieldName))
							g.P()

							reqField = fmt.Sprintf("%s: %s, \n", generator.CamelCase(fieldName), fieldName)
						}

						arg += reqField
					}

					req := fmt.Sprintf("req := &%s{%s}", inputTypeName, arg)
					g.P(req)
					g.P()

					g.P(fmt.Sprintf("rsp, err := h.%sHandler.%s(%s)", servName, method.GetName(), "req"))
					g.P("if err != nil {")
					g.P("h.fail(c, err)")
					g.P("return")
					g.P("}")
					g.P()
					g.P("h.returnBack(c, err, rsp)")
				}
			} else {
				// get 请求无参数
				g.P(fmt.Sprintf("rsp, err := h.%sHandler.%s()", servName, method.GetName()))
				g.P("if err != nil {")
				g.P("h.fail(c, err)")
				g.P("return")
				g.P("}")
				g.P()
				g.P("h.returnBack(c, err, rsp)")
			}

			g.P("}")
			g.P()
		}
		g.P()
	}

	// fail
	g.P(fmt.Sprintf("func (h *%s_Router) fail(c *gin.Context, err error) {", servName))
	g.P("c.JSON(http.StatusBadRequest, gin.H{\"code\": 1, \"msg\": err.Error(), \"data\": \"\"})")
	g.P("}")
	g.P()

	// success
	g.P(fmt.Sprintf("func (h *%s_Router) success(c *gin.Context, data interface{}) {", servName))
	g.P("c.JSON(http.StatusOK, gin.H{\"code\": 0, \"msg\": \"\", \"data\": data})")
	g.P("}")
	g.P()

	// return back
	g.P(fmt.Sprintf("func (h *%s_Router) returnBack(c *gin.Context, err error, data interface{}) {", servName))
	g.P("if err != nil {")
	g.P("h.fail(c, err)")
	g.P("} else {")
	g.P("h.success(c, data)")
	g.P("}")
	g.P("}")
	g.P()
}

func (g *pms) getRule(method *descriptor.MethodDescriptorProto) (string, string) {
	if method.Options == nil || !proto.HasExtension(method.Options, options.E_Http) {
		return "", ""
	}
	// http rules
	r := proto.GetExtension(method.Options, options.E_Http)
	// if err != nil {
	// 	return "", ""
	// }

	rule := r.(*options.HttpRule)
	var meth string
	var path string
	switch {
	case len(rule.GetDelete()) > 0:
		meth = "DELETE"
		path = rule.GetDelete()
	case len(rule.GetGet()) > 0:
		meth = "GET"
		path = rule.GetGet()
	case len(rule.GetPatch()) > 0:
		meth = "PATCH"
		path = rule.GetPatch()
	case len(rule.GetPost()) > 0:
		meth = "POST"
		path = rule.GetPost()
	case len(rule.GetPut()) > 0:
		meth = "PUT"
		path = rule.GetPut()
	}

	if len(meth) == 0 || len(path) == 0 {
		return "", ""
	}

	return meth, path
}

func typeEmpty(typeName string) bool {
	return typeName == "emptypb.Empty"
}
