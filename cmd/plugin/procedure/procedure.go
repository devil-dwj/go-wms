package procedure

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/devil-dwj/go-wms/cmd/generator"

	"google.golang.org/protobuf/types/descriptorpb"
)

func init() {
	generator.RegisterPlugin(new(Procedure))
}

type Procedure struct {
	gen *generator.Generator
}

func (g *Procedure) Name() string {
	return "Procedure"
}

func (g *Procedure) Init(gen *generator.Generator) {
	g.gen = gen
}

func (g *Procedure) objectNamed(name string) generator.Object {
	g.gen.RecordTypeUse(name)
	return g.gen.ObjectNamed(name)
}

func (g *Procedure) typeName(str string) string {
	return g.gen.TypeName(g.objectNamed(str))
}

func (g *Procedure) P(args ...interface{}) { g.gen.P(args...) }

func (g *Procedure) Generate(file *generator.FileDescriptor) {

	origFileName := file.GetName()
	moduleName := strings.TrimSuffix(origFileName, "_pr.proto")
	servName := generator.CamelCase(moduleName)

	// interface
	g.P(fmt.Sprintf("type %sProcedure interface {", servName))
	g.P("GetRawDB() *gorm.DB")

	for _, message := range file.FileDescriptorProto.MessageType {
		if !strings.HasPrefix(message.GetName(), "PrPs") {
			continue
		}

		var (
			argInput     = ""
			argOut       = ""
			argOutResult = ""
			argTotal     = ""
		)

		for _, field := range message.GetField() {
			typeName := ""
			if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
				typeName = g.typeName(field.GetTypeName())
			} else {
				typeName = field.Type.String()
			}

			if field.GetName() == "req" {
				argInput = fmt.Sprintf("req *%s", typeName)
			}
			if field.GetName() == "rsp" {
				argOut = fmt.Sprintf("[]*%s, ", typeName)
			}
			if field.GetName() == "result" {
				argOutResult = fmt.Sprintf("%s, ", "int")
			}
			if field.GetName() == "total_count" {
				argTotal = fmt.Sprintf("%s, ", "int")
			}
		}
		g.P(fmt.Sprintf("%s(%s) (%s%s%serror)",
			message.GetName(), argInput, argOut, argOutResult, argTotal))
	}
	g.P("}")

	// Procedure struct
	g.P(fmt.Sprintf("type %s_Procedure struct {", servName))
	g.P("db *gorm.DB")
	g.P("}")

	// new Proceduresitory struct
	g.P(fmt.Sprintf("func New%sProcedure(db *gorm.DB) %sProcedure {", servName, servName))
	g.P(fmt.Sprintf("return &%s_Procedure{", servName))
	g.P("db: db,")
	g.P("}")
	g.P("}")
	g.P()

	// raw db
	g.P(fmt.Sprintf("func (r *%s_Procedure) GetRawDB() *gorm.DB {", servName))
	g.P("return r.db")
	g.P("}")
	g.P()

	for _, message := range file.FileDescriptorProto.MessageType {
		if !strings.HasPrefix(message.GetName(), "PrPs") {
			continue
		}

		var (
			argInput = ""

			argOut       = ""
			argOutResult = ""
			argOutTotal  = ""

			resIdent    = ""
			resultIdent = ""
			totalIdent  = ""

			returnRes    = ""
			returnResult = ""
			returnTotal  = ""
		)

		for _, field := range message.GetField() {
			typeName := ""
			if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
				typeName = g.typeName(field.GetTypeName())
			} else {
				typeName = field.Type.String()
			}

			if field.GetName() == "req" {
				argInput = fmt.Sprintf("req *%s", typeName)
			}

			if field.GetName() == "rsp" {
				argOut = fmt.Sprintf("[]*%s, ", typeName)
				resIdent = fmt.Sprintf("res := []*%s{}", typeName)
				returnRes = "res, "
			}

			if field.GetName() == "result" {
				argOutResult = fmt.Sprintf("%s, ", "int")
				resultIdent = "result := -1"
				returnResult = "result, "
			}

			if field.GetName() == "total_count" {
				argOutTotal = fmt.Sprintf("%s, ", "int")
				totalIdent = "total := 0"
				returnTotal = "total, "
			}
		}

		// total_count

		// func
		g.P(fmt.Sprintf("func (r *%s_Procedure) %s(%s) (%s%s%serror) {",
			servName, message.GetName(), argInput, argOut, argOutResult, argOutTotal))
		g.P(resIdent)
		g.P(resultIdent)
		g.P(totalIdent)
		g.P("err := r.db.")

		// iv field
		ivTempl := ""
		ivTotal := ""
		ivArg := ""
		ivSp := ""
		for _, field := range message.GetField() {
			if field.GetName() == "req" {
				ivSp = ","

				fieldTypeName := g.typeName(field.GetTypeName())
				for _, message1 := range file.FileDescriptorProto.MessageType {
					if message1.GetName() == fieldTypeName {
						ide := ", "
						fieldLen := len(message1.GetField())
						for i, fieldReq := range message1.GetField() {
							if i == fieldLen-1 {
								ide = ""
							}

							ivTempl += ", ?"
							ivArg += fmt.Sprintf("req.%s%s",
								generator.CamelCase(fieldReq.GetName()), ide)
						}
					}
				}
			}

			if field.GetName() == "total_count" {
				ivTotal += ", @ov_total_count"
			}
		}

		if ivTempl != "" {
			// 有入参
			g.P(fmt.Sprintf(`Raw("call %s(@ov_return%s%s)"%s`,
				g.Camel2Case(message.GetName()), ivTempl, ivTotal, ivSp))
			g.P(fmt.Sprintf("%s).", ivArg))
			if resIdent != "" {
				g.P("Scan(&res).")
			} else {
				if resultIdent == "" {
					g.P("result := -1")
				}
				g.P("Scan(&result).")
			}

			if resultIdent != "" {
				g.P(`Raw("select @ov_return").`)
				g.P("Scan(&result).")
			}
			if totalIdent != "" {
				g.P(`Raw("select @ov_total_count").`)
				g.P("Scan(&total).")
			}

		} else {
			// 无入参
			g.P(fmt.Sprintf(`Raw("call %s(@ov_return%s)").`,
				g.Camel2Case(message.GetName()), ivTotal))
			if resIdent != "" {
				g.P("Scan(&res).")
			}
			if resultIdent != "" {
				g.P(`Raw("select @ov_return").`)
				g.P("Scan(&result).")
			}
			if totalIdent != "" {
				g.P(`Raw("select @ov_total_count").`)
				g.P("Scan(&total).")
			}
		}

		g.P("Error")
		g.P()

		g.P(fmt.Sprintf("return %s%s%serr", returnRes, returnResult, returnTotal))
		g.P("}")
		g.P()
	}
}

func (g *Procedure) Camel2Case(name string) string {
	buffer := NewBuffer()
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i != 0 {
				buffer.Append('_')
			}
			buffer.Append(unicode.ToLower(r))
		} else {
			buffer.Append(r)
		}
	}
	return buffer.String()
}

type Buffer struct {
	*bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{Buffer: new(bytes.Buffer)}
}

func (b *Buffer) Append(i interface{}) *Buffer {
	switch val := i.(type) {
	case int:
		b.append(strconv.Itoa(val))
	case int64:
		b.append(strconv.FormatInt(val, 10))
	case uint:
		b.append(strconv.FormatUint(uint64(val), 10))
	case uint64:
		b.append(strconv.FormatUint(val, 10))
	case string:
		b.append(val)
	case []byte:
		b.Write(val)
	case rune:
		b.WriteRune(val)
	}
	return b
}

func (b *Buffer) append(s string) *Buffer {
	b.WriteString(s)
	return b
}

func (g *Procedure) GenerateImports(file *generator.FileDescriptor, imports map[generator.GoImportPath]generator.GoPackageName) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}

}
