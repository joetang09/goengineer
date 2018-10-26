# goengineer


This is an app engineer simply.

Use components by implement engineer/Component.
Register a server by implement engineer/Server.

Whatever what config is using by your component, the engineer will correct parse it. 
And the raw package has not injury.

But there is some of flaws clearly, wuwuwuwu...

But, It is useful to you sometimes, Because it is in running.

Packing your component just like this : 

```go

type Config struct {
    List    string
    Your    int8
    Fields  []int
    Is      bool
    Ok      struct{}
}

type Cpnt struct {
}

func (Cpnt) Init(ops ...interface{}) error { // config will pass to ops[0]
	return nil
}

func (Cpnt) CfgKey() string { // fetch what key to parse to your Config
	return "awesomekey"
}

func (Cpnt) CfgType() interface{} { // return you config value
	return Config{}
}

func (Cpnt) CfgUpdate(interface{}) { // it's useless

}

```


just like this :

```go

engineer.BuildEnv(true) // parsing config if true

enginer.Use(db.Cpnt{}) // use component like this

ws := webserver.WebServer{}
enginer.Use(ws)
enginer.RegisterServer(ws) // register server like this

engineer.Start() // start your app like this
```