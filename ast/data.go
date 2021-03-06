package ast

// Module represents the namespace of a program.
// Helps segregate different components and avoid identifier conflicts
// Represented as Root / Sub or just Root if Sub is not provided
type Module struct {
	Root Identifier
	Sub  Identifier
}

func (m *Module) ToString() string {
	if &m.Sub != nil {
		return m.Root.name
	}
	return m.Root.name + "/" + m.Sub.name
}

type Parameter struct {
	Type       Type
	Identifier Identifier
}

func (p *Parameter) ToString() string {
	res := ""
	if p.Type != nil {
		res += p.Type.ToString() + " "
	}
	res += p.Identifier.name
	return res
}

type Entry struct {
	Key   Expression
	Value Expression
}

func (p *Entry) ToString() string {
	return "(" + p.Key.ToString() + " : " + p.Value.ToString() + ")"
}

type StructField struct {
	Type       Type
	Identifier Identifier
}

func (p *StructField) ToString() string {
	return p.Type.ToString() + " " + p.Identifier.name
}
