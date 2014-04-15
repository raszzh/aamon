package main

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"time"
	"net/http"
	"io/ioutil"
	"net/smtp"
	"os"
	"log"
	"encoding/json"
)

type Configuration struct {
	Mysqlparams string
	Smtphost string
	Smtpuser string
	Smtppass string
	Sender string
	Smtpport uint
	Timeout uint
	Interval uint
}


var bbstatus map[string] time.Time

func checkstatus(ip string) bool {
	resp,err:=http.Get("http://"+ip)
	if err!=nil {
		return false
	}
	defer resp.Body.Close()
	body,err:=ioutil.ReadAll(resp.Body)
	if err!=nil { return false }
	s:=string(body[:8])
	return "It works"==s
}

func sendmail(ip,to,from string) {
	tolist:=[]string{"rashost@qq.com"}
	template:=`From: %s
To: %s
Subject: %s
MIME_version: 1.0
Content-Type: text/plain; charset="UTF-8"
	
Hello,

%s	

Regards,
`
	subject:=fmt.Sprintf("IP %s fails from %s to %s",ip,from,to)
	body:=fmt.Sprintf(template,conf.Sender,"rashost@qq.com",subject,subject)
	auth:=smtp.PlainAuth("",conf.Smtpuser,conf.Smtppass,conf.Smtphost)
	smtp.SendMail(fmt.Sprintf("%s:%d",conf.Smtphost,conf.Smtpport),auth,conf.Sender,tolist,[]byte(body))
}

func bbcheck(ip string) {
	_,ok:=bbstatus[ip]
	if !ok {
		bbstatus[ip]=time.Time{}
	}
	good:=checkstatus(ip)
	log.Println(ip,good)
	if good {
		bbstatus[ip]=time.Time{}
	}else if bbstatus[ip].IsZero() {
		bbstatus[ip]=time.Now()
	}else if time.Now().After( bbstatus[ip].Add( time.Minute*time.Duration(conf.Timeout)) )  {
		sendmail(ip,time.Now().String(),bbstatus[ip].String())
		bbstatus[ip]=time.Now()
	}else {
			//
	}

	
}

func work() {
	db, err := sql.Open("mysql", conf.Mysqlparams)
	if err != nil {  
		log.Println(err)//仅仅是显示异常
		return
	}
	defer db.Close()  //只有在前面用了 panic 这时defer才能起作用
	rows, err := db.Query("select ip from BB where status=1") //从新闻表取出两个字段
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	var ip string
	for rows.Next() {
		rerr:=rows.Scan(&ip)
		if rerr==nil {
			bbcheck(ip)
		}
	}
	log.Println()
}
var conf Configuration

func main() {

	file,err:=os.Open("/etc/aamon.json")
	if err != nil { panic(err) }
	defer func() { 
		err=file.Close()
		if err!=nil { panic(err) }
	}()
	decoder:=json.NewDecoder(file)
	err=decoder.Decode(&conf)
	if err!=nil { panic(err) }

	fmt.Println(&conf)

	bbstatus=make(map[string]time.Time)

	log.Printf("Hello aamon!\n")
	for true {
		work()
		time.Sleep( time.Duration(conf.Interval)*time.Minute )
	}
}

