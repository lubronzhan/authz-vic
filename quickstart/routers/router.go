package routers

import (
	"github.com/lubronzhan/authz-vic/quickstart/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
}
