package main
import(
    "login"
)

func main(){
    l := new(login.Login)
    l.Connect("../sql/user.db")
    // call l.NewAcount() to create a new account
    // l.NewAcount(id, password, name)
}
