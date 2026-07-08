package board

type Column int

const (
    _ Column = iota
    
    ColumnId
    ColumnTitle
    ColumnContent
    ColumnImg
    ColumnUser
    ColumnDate
)

type Params struct {
    Column Column
    Value interface{}
}




