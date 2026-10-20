package rest


import (
	"basicproject_go/controllers"
	"basicproject_go/global/jwt"
	"basicproject_go/models"

    "regexp"
    "strings"
)

type UserController struct {
	controllers.Controller
}

func (c *UserController) Read(id int64) {
    
    
	conn := c.NewConnection()

	manager := models.NewUserManager(conn)
	item := manager.Get(id)

    
    
    c.Set("item", item)
}

func (c *UserController) Index(page int, pagesize int) {
    
    
	conn := c.NewConnection()

	manager := models.NewUserManager(conn)

    var args []interface{}
    
    _email := c.Get("email")
    if _email != "" {
        args = append(args, models.Where{Column:"email", Value:_email, Compare:"like"})
    }
    _name := c.Get("name")
    if _name != "" {
        args = append(args, models.Where{Column:"name", Value:_name, Compare:"="})
        
    }
    _passwd := c.Get("passwd")
    if _passwd != "" {
        args = append(args, models.Where{Column:"passwd", Value:_passwd, Compare:"like"})
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
                    str += ", u_" + strings.Trim(v, " ")                
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

func (c *UserController) Count() {
    
    
	conn := c.NewConnection()

	manager := models.NewUserManager(conn)

    var args []interface{}
    
    _email := c.Get("email")
    if _email != "" {
        args = append(args, models.Where{Column:"email", Value:_email, Compare:"like"})
        
    }
    _name := c.Get("name")
    if _name != "" {
        args = append(args, models.Where{Column:"name", Value:_name, Compare:"="})
        
        
    }
    _passwd := c.Get("passwd")
    if _passwd != "" {
        args = append(args, models.Where{Column:"passwd", Value:_passwd, Compare:"like"})
        
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

func (c *UserController) Insert(item *models.UserUpdate) {
    
    
    
    if item.Passwd != "" {
        hashed, err := jwt.GeneratePasswd(item.Passwd)
        if err == nil {
            item.Passwd = hashed
        }
    }
    

	conn := c.NewConnection()
    
	manager := models.NewUserManager(conn)
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

func (c *UserController) Insertbatch(item *[]models.UserUpdate) {  
    if item == nil || len(*item) == 0 {
        return
    }

    rows := len(*item)
    
    
    
	conn := c.NewConnection()
    
	manager := models.NewUserManager(conn)

    for i := 0; i < rows; i++ {
        
        if (*item)[i].Passwd != "" {
            hashed, err := jwt.GeneratePasswd((*item)[i].Passwd)
            if err == nil {
                (*item)[i].Passwd = hashed
            }
        }
        
	    err := manager.Insert(&((*item)[i]))
        if err != nil {
            c.Set("code", "error")    
            c.Set("error", err)
            return
        }
    }
}

func (c *UserController) Update(item *models.UserUpdate) {
    
    
    
    if item.Passwd != "" {
        hashed, err := jwt.GeneratePasswd(item.Passwd)
        if err == nil {
            item.Passwd = hashed
        }
    }
    

	conn := c.NewConnection()

	manager := models.NewUserManager(conn)
    err := manager.Update(item)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
        return
    }
}

func (c *UserController) Delete(item *models.User) {
    
    
    conn := c.NewConnection()

	manager := models.NewUserManager(conn)

    
	err := manager.Delete(item.Id)
    if err != nil {
        c.Set("code", "error")    
        c.Set("error", err)
    }
}

func (c *UserController) Deletebatch(item *[]models.User) {
    
    
    conn := c.NewConnection()

	manager := models.NewUserManager(conn)

    for _, v := range *item {
        
    
	    err := manager.Delete(v.Id)
        if err != nil {
            c.Set("code", "error")    
            c.Set("error", err)
            return
        }
    }
}


