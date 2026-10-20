package rest


import (
	"basicproject_go/controllers"
	
	"basicproject_go/models"

    "regexp"
    "strings"
)

type BoardController struct {
	controllers.Controller
}

func (c *BoardController) Read(id int64) {
    
    
	conn := c.NewConnection()

	manager := models.NewBoardManager(conn)
	item := manager.Get(id)

    
    
    c.Set("item", item)
}

func (c *BoardController) Index(page int, pagesize int) {
    
    
	conn := c.NewConnection()

	manager := models.NewBoardManager(conn)

    var args []interface{}
    
    _title := c.Get("title")
    if _title != "" {
        args = append(args, models.Where{Column:"title", Value:_title, Compare:"="})
        
    }
    _content := c.Get("content")
    if _content != "" {
        args = append(args, models.Where{Column:"content", Value:_content, Compare:"="})
        
    }
    _img := c.Get("img")
    if _img != "" {
        args = append(args, models.Where{Column:"img", Value:_img, Compare:"like"})
    }
    _user := c.Geti64("user")
    if _user != 0 {
        args = append(args, models.Where{Column:"user", Value:_user, Compare:"="})    
    }
    _startdate := c.Get("startdate")
    _enddate := c.Get("enddate")
    if _startdate != "" && _enddate != "" {        
        var v [2]string
        v[0] = _startdate
        v[1] = _enddate  
        args = append(args, models.Where{Column:"date", Value:v, Compare:"between"})    
    } else if  _startdate != "" {          
        args = append(args, models.Where{Column:"date", Value:_startdate, Compare:">="})
    } else if  _enddate != "" {          
        args = append(args, models.Where{Column:"date", Value:_enddate, Compare:"<="})            
    }
    

    
    
    if page != 0 && pagesize != 0 {
        args = append(args, models.Paging(page, pagesize))
    }
    
    orderby := c.Get("orderby")
    if orderby == "" {
        if page != 0 && pagesize != 0 {
            orderby = "id desc"
            args = append(args, models.Ordering(orderby))
        }
    } else {
        orderbys := strings.Split(orderby, ",")

        str := ""
        for i, v := range orderbys {
            if i == 0 {
                str += v
            } else {
                if strings.Contains(v, "_") {                   
                    str += ", " + strings.Trim(v, " ")
                } else {
                    str += ", b_" + strings.Trim(v, " ")                
                }
            }
        }
        
        // orderby 는 ORDER BY 절에 그대로 들어가므로 "컬럼 [asc|desc]" 목록 형태만 허용한다 (SQL injection 방지).
        if regexp.MustCompile(`^\s*[A-Za-z0-9_]+(\s+(?i:asc|desc))?\s*(,\s*[A-Za-z0-9_]+(\s+(?i:asc|desc))?\s*)*$`).MatchString(orderby) {
        	args = append(args, models.Ordering(str))
        }
    }
    
	items := manager.Find(args)
	c.Set("items", items)

    if page == 1 {
       total := manager.Count(args)
	   c.Set("total", total)
    }
}

func (c *BoardController) Count() {
    
    
	conn := c.NewConnection()

	manager := models.NewBoardManager(conn)

    var args []interface{}
    
    _title := c.Get("title")
    if _title != "" {
        args = append(args, models.Where{Column:"title", Value:_title, Compare:"="})
        
        
    }
    _content := c.Get("content")
    if _content != "" {
        args = append(args, models.Where{Column:"content", Value:_content, Compare:"="})
        
        
    }
    _img := c.Get("img")
    if _img != "" {
        args = append(args, models.Where{Column:"img", Value:_img, Compare:"like"})
        
    }
    _user := c.Geti64("user")
    if _user != 0 {
        args = append(args, models.Where{Column:"user", Value:_user, Compare:"="})    
    }
    _startdate := c.Get("startdate")
    _enddate := c.Get("enddate")

    if _startdate != "" && _enddate != "" {        
        var v [2]string
        v[0] = _startdate
        v[1] = _enddate  
        args = append(args, models.Where{Column:"date", Value:v, Compare:"between"})    
    } else if  _startdate != "" {          
        args = append(args, models.Where{Column:"date", Value:_startdate, Compare:">="})
    } else if  _enddate != "" {          
        args = append(args, models.Where{Column:"date", Value:_enddate, Compare:"<="})            
    }
    
    
    
    
    total := manager.Count(args)
	c.Set("total", total)
}

func (c *BoardController) Insert(item *models.Board) {
    
    
    

	conn := c.NewConnection()
    
	manager := models.NewBoardManager(conn)
	err := manager.Insert(item)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
        return
    }

    id := manager.GetIdentity()
    c.Result["id"] = id
    item.Id = id
}

func (c *BoardController) Insertbatch(item *[]models.Board) {  
    if item == nil || len(*item) == 0 {
        return
    }

    rows := len(*item)
    
    
    
	conn := c.NewConnection()
    
	manager := models.NewBoardManager(conn)

    for i := 0; i < rows; i++ {
        
	    err := manager.Insert(&((*item)[i]))
        if err != nil {
            c.Set("code", "error")    
            c.Set("error", err)
            return
        }
    }
}

func (c *BoardController) Update(item *models.Board) {
    
    
    

	conn := c.NewConnection()

	manager := models.NewBoardManager(conn)
    err := manager.Update(item)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
        return
    }
}

func (c *BoardController) Delete(item *models.Board) {
    
    
    conn := c.NewConnection()

	manager := models.NewBoardManager(conn)

    
	err := manager.Delete(item.Id)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
    }
}

func (c *BoardController) Deletebatch(item *[]models.Board) {
    
    
    conn := c.NewConnection()

	manager := models.NewBoardManager(conn)

    for _, v := range *item {
        
    
	    err := manager.Delete(v.Id)
        if err != nil {
            c.Set("code", "error")    
            c.Set("error", err)
            return
        }
    }
}


