package compiler

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"go-java-compiler/src/ast"
	"go-java-compiler/src/lexer"
)

// ConstantPool manages the constant pool in a JVM class file
type ConstantPool struct {
	entries []interface{} // index 0 is unused, starts at 1
	lookup  map[string]int
}

func NewConstantPool() *ConstantPool {
	return &ConstantPool{
		entries: []interface{}{nil}, // index 0 is reserved
		lookup:  make(map[string]int),
	}
}

type CPEntry interface{}

type CPUtf8 struct {
	value string
}

type CPInteger struct {
	value int32
}

type CPClass struct {
	nameIndex int
}

type CPString struct {
	stringIndex int
}

type CPFieldref struct {
	classIndex       int
	nameAndTypeIndex int
}

type CPMethodref struct {
	classIndex       int
	nameAndTypeIndex int
}

type CPNameAndType struct {
	nameIndex       int
	descriptorIndex int
}

func (cp *ConstantPool) addUtf8(s string) int {
	key := "utf8:" + s
	if idx, exists := cp.lookup[key]; exists {
		return idx
	}
	cp.entries = append(cp.entries, CPUtf8{value: s})
	idx := len(cp.entries) - 1
	cp.lookup[key] = idx
	return idx
}

func (cp *ConstantPool) addClass(name string) int {
	nameIdx := cp.addUtf8(name)
	key := "class:" + name
	if idx, exists := cp.lookup[key]; exists {
		return idx
	}
	cp.entries = append(cp.entries, CPClass{nameIndex: nameIdx})
	idx := len(cp.entries) - 1
	cp.lookup[key] = idx
	return idx
}

func (cp *ConstantPool) addString(s string) int {
	strIdx := cp.addUtf8(s)
	key := "string:" + s
	if idx, exists := cp.lookup[key]; exists {
		return idx
	}
	cp.entries = append(cp.entries, CPString{stringIndex: strIdx})
	idx := len(cp.entries) - 1
	cp.lookup[key] = idx
	return idx
}

func (cp *ConstantPool) addMethodref(className, methodName, descriptor string) int {
	classIdx := cp.addClass(className)
	nameIdx := cp.addUtf8(methodName)
	descIdx := cp.addUtf8(descriptor)
	key := "methodref:" + className + "." + methodName + descriptor
	if idx, exists := cp.lookup[key]; exists {
		return idx
	}

	// Add NameAndType entry
	cp.entries = append(cp.entries, CPNameAndType{nameIndex: nameIdx, descriptorIndex: descIdx})
	natIdx := len(cp.entries) - 1
	
	// Add Methodref entry
	cp.entries = append(cp.entries, CPMethodref{classIndex: classIdx, nameAndTypeIndex: natIdx})
	methodrefIdx := len(cp.entries) - 1

	cp.lookup[key] = methodrefIdx
	return methodrefIdx
}

func (cp *ConstantPool) addInteger(val int32) int {
	key := fmt.Sprintf("int:%d", val)
	if idx, exists := cp.lookup[key]; exists {
		return idx
	}
	cp.entries = append(cp.entries, CPInteger{value: val})
	idx := len(cp.entries) - 1
	cp.lookup[key] = idx
	return idx
}

func (cp *ConstantPool) write(buf *bytes.Buffer) {
	// Write constant pool count
	binary.Write(buf, binary.BigEndian, uint16(len(cp.entries)))

	// Write each constant pool entry
	for i := 1; i < len(cp.entries); i++ {
		entry := cp.entries[i]
		switch e := entry.(type) {
		case CPUtf8:
			buf.WriteByte(1) // CONSTANT_Utf8
			binary.Write(buf, binary.BigEndian, uint16(len(e.value)))
			buf.WriteString(e.value)
		case CPInteger:
			buf.WriteByte(3) // CONSTANT_Integer
			binary.Write(buf, binary.BigEndian, e.value)
		case CPClass:
			buf.WriteByte(7) // CONSTANT_Class
			binary.Write(buf, binary.BigEndian, uint16(e.nameIndex))
		case CPString:
			buf.WriteByte(8) // CONSTANT_String
			binary.Write(buf, binary.BigEndian, uint16(e.stringIndex))
		case CPMethodref:
			buf.WriteByte(10) // CONSTANT_Methodref
			binary.Write(buf, binary.BigEndian, uint16(e.classIndex))
			binary.Write(buf, binary.BigEndian, uint16(e.nameAndTypeIndex))
		case CPNameAndType:
			buf.WriteByte(12) // CONSTANT_NameAndType
			binary.Write(buf, binary.BigEndian, uint16(e.nameIndex))
			binary.Write(buf, binary.BigEndian, uint16(e.descriptorIndex))
		}
	}
}

// MethodCodeGenerator generates bytecode for a method
type MethodCodeGenerator struct {
	code            []byte
	localVarCount   int
	symTable        map[string]int // variable name -> local variable index
	maxStack        int
	currentStack    int
	cp              *ConstantPool
}

func NewMethodCodeGenerator(cp *ConstantPool) *MethodCodeGenerator {
	return &MethodCodeGenerator{
		code:      []byte{},
		symTable:  make(map[string]int),
		maxStack:  10,
		cp:        cp,
	}
}

func (m *MethodCodeGenerator) emit(opcode byte, args ...byte) {
	m.code = append(m.code, opcode)
	m.code = append(m.code, args...)
}

func (m *MethodCodeGenerator) emitU2(val uint16) {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, val)
	m.code = append(m.code, buf.Bytes()...)
}

func (m *MethodCodeGenerator) declareVar(name string) {
	m.symTable[name] = m.localVarCount
	m.localVarCount++
}

func (m *MethodCodeGenerator) getVarIndex(name string) int {
	return m.symTable[name]
}

func (m *MethodCodeGenerator) generateExpr(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		if e.Value >= -1 && e.Value <= 5 {
			m.emit(byte(ICONST_0 + e.Value))
		} else if e.Value >= -128 && e.Value <= 127 {
			m.emit(BIPUSH, byte(e.Value))
		} else {
			// Use LDC with integer constant
			idx := m.cp.addInteger(int32(e.Value))
			m.emit(LDC, byte(idx))
		}
	case *ast.Identifier:
		idx := m.getVarIndex(e.Name)
		// Use appropriate ILOAD variant
		if idx >= 0 && idx <= 3 {
			m.emit(ILOAD_0 + byte(idx))
		} else {
			m.emit(ILOAD, byte(idx))
		}
	case *ast.BinaryExpr:
		m.generateExpr(e.Left)
		m.generateExpr(e.Right)
		switch e.Op {
		case lexer.PLUS:
			m.emit(IADD)
		case lexer.MINUS:
			m.emit(ISUB)
		case lexer.ASTERISK:
			m.emit(IMUL)
		case lexer.SLASH:
			m.emit(IDIV)
		}
	case *ast.StringLiteral:
		idx := m.cp.addString(e.Value)
		m.emit(LDC, byte(idx))
	case *ast.MethodCall:
		if ident, ok := e.Object.(*ast.Identifier); ok && ident.Name == "System" && e.Method == "out" {
			// Skip - this is a field access
		} else if ident, ok := e.Object.(*ast.Identifier); ok && ident.Name == "out" && e.Method == "println" {
			// System.out.println call
			for _, arg := range e.Arguments {
				m.generateExpr(arg)
			}
			// call System.out.println
			idx := m.cp.addMethodref("java/io/PrintStream", "println", "(I)V")
			m.emit(INVOKEVIRTUAL)
			m.emitU2(uint16(idx))
		}
	}
}

// JVM Bytecode Constants
const (
	ICONST_0      = 0x03
	BIPUSH        = 0x10
	LDC           = 0x12
	ILOAD         = 0x15
	ILOAD_0       = 0x1A
	ISTORE        = 0x36
	ISTORE_0      = 0x3B
	IADD          = 0x60
	ISUB          = 0x64
	IMUL          = 0x68
	IDIV          = 0x6C
	IRETURN       = 0xAC
	RETURN        = 0xB1
	INVOKEVIRTUAL = 0xB6
	INVOKESTATIC  = 0xB8
)

// Compiler generates a complete class file
type Compiler struct {
	program *ast.Program
	cp      *ConstantPool
}

func NewCompiler(program *ast.Program) *Compiler {
	return &Compiler{
		program: program,
		cp:      NewConstantPool(),
	}
}

func (c *Compiler) Compile() ([]byte, error) {
	if len(c.program.Classes) == 0 {
		return nil, fmt.Errorf("no classes to compile")
	}

	classDef := c.program.Classes[0]
	return c.compileClass(classDef)
}

func (c *Compiler) compileClass(class *ast.ClassDecl) ([]byte, error) {
	if class.Name == "" {
		return nil, fmt.Errorf("class has no name")
	}
	if len(class.Methods) == 0 {
		return nil, fmt.Errorf("class %s has no methods", class.Name)
	}

	buf := &bytes.Buffer{}

	// Magic number
	binary.Write(buf, binary.BigEndian, uint32(0xCAFEBABE))

	// Version
	binary.Write(buf, binary.BigEndian, uint16(0)) // minor version
	binary.Write(buf, binary.BigEndian, uint16(52)) // major version (Java 8)

	// Constant pool
	c.cp.addClass("java/lang/Object")
	c.cp.addClass(class.Name)
	c.cp.addUtf8("Code")
	c.cp.addUtf8("LineNumberTable")

	for _, method := range class.Methods {
		c.cp.addUtf8(method.Name)
		c.cp.addUtf8(c.getMethodDescriptor(method))
	}

	c.cp.write(buf)

	// Access flags (public)
	binary.Write(buf, binary.BigEndian, uint16(0x0001))

	// This class
	thisClassIdx := c.cp.addClass(class.Name)
	binary.Write(buf, binary.BigEndian, uint16(thisClassIdx))

	// Super class (Object)
	superClassIdx := c.cp.addClass("java/lang/Object")
	binary.Write(buf, binary.BigEndian, uint16(superClassIdx))

	// Interfaces
	binary.Write(buf, binary.BigEndian, uint16(0))

	// Fields
	binary.Write(buf, binary.BigEndian, uint16(0))

	// Methods
	binary.Write(buf, binary.BigEndian, uint16(len(class.Methods)))

	for _, method := range class.Methods {
		c.compileMethod(buf, method, class.Name)
	}

	// Attributes
	binary.Write(buf, binary.BigEndian, uint16(0))

	return buf.Bytes(), nil
}

func (c *Compiler) compileMethod(buf *bytes.Buffer, method *ast.MethodDecl, className string) {
	// Access flags
	flags := uint16(0x0009) // public static
	if !method.IsStatic {
		flags = 0x0001 // public
	}
	binary.Write(buf, binary.BigEndian, flags)

	// Name index
	nameIdx := c.cp.addUtf8(method.Name)
	binary.Write(buf, binary.BigEndian, uint16(nameIdx))

	// Descriptor index
	descIdx := c.cp.addUtf8(c.getMethodDescriptor(method))
	binary.Write(buf, binary.BigEndian, uint16(descIdx))

	// Attributes count
	binary.Write(buf, binary.BigEndian, uint16(1))

	// Code attribute
	codeAttrNameIdx := c.cp.addUtf8("Code")
	binary.Write(buf, binary.BigEndian, uint16(codeAttrNameIdx))

	// Generate bytecode
	codeGen := NewMethodCodeGenerator(c.cp)

	// Process methods body
	for _, stmt := range method.Body {
		switch s := stmt.(type) {
		case *ast.VarDecl:
			codeGen.declareVar(s.Name)
			if s.Value != nil {
				codeGen.generateExpr(s.Value)
				idx := codeGen.getVarIndex(s.Name)
				// Use appropriate ISTORE variant
				if idx >= 0 && idx <= 3 {
					codeGen.emit(ISTORE_0 + byte(idx))
				} else {
					codeGen.emit(ISTORE, byte(idx))
				}
			}
		case *ast.ExpressionStatement:
			codeGen.generateExpr(s.Expr)
		case *ast.ReturnStatement:
			if s.Value != nil {
				codeGen.generateExpr(s.Value)
				codeGen.emit(IRETURN)
			} else {
				codeGen.emit(RETURN)
			}
		}
	}

	// Add implicit return if none exists
	if len(codeGen.code) == 0 || codeGen.code[len(codeGen.code)-1] != RETURN {
		codeGen.emit(RETURN)
	}

	// Write code attribute length
	codeLen := 12 + len(codeGen.code) // 12 for max_stack,max_locals,code_length + code
	binary.Write(buf, binary.BigEndian, uint32(codeLen))

	// Max stack
	binary.Write(buf, binary.BigEndian, uint16(10))

	// Max locals
	binary.Write(buf, binary.BigEndian, uint16(codeGen.localVarCount + 1))

	// Code length
	binary.Write(buf, binary.BigEndian, uint32(len(codeGen.code)))

	// Code
	buf.Write(codeGen.code)

	// Exception table length
	binary.Write(buf, binary.BigEndian, uint16(0))

	// Code attributes (LineNumberTable, etc.)
	binary.Write(buf, binary.BigEndian, uint16(0))
}

func (c *Compiler) getMethodDescriptor(method *ast.MethodDecl) string {
	desc := "("
	for _, param := range method.Params {
		if param.Type == "int" {
			desc += "I"
		} else if param.Type == "String[]" {
			desc += "[Ljava/lang/String;"
		} else if param.Type == "String" {
			desc += "Ljava/lang/String;"
		} else {
			desc += "I" // default to int
		}
	}
	desc += ")"
	if method.ReturnType == "int" {
		desc += "I"
	} else {
		desc += "V"
	}
	return desc
}
