package user

type Column int

const (
    _ Column = iota
    
    ColumnId
    ColumnEmail
    ColumnName
    ColumnPasswd
    ColumnDate
)

type Params struct {
    Column Column
    Value interface{}
}




