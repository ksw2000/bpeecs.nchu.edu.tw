package main
import(
    "bpeecs.nchu.edu.tw/login"
)

func main(){
    l := new(login.Login)
    l.Connect("./db/main.db")
    // call l.NewAcount() to create a new account
    // l.NewAcount(id, password, name)
}
