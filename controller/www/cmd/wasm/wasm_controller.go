package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"syscall/js"
	"time"
)

// GOOS=js GOARCH=wasm go build -o ../../assets/json.wasm

type WebError struct {
	err string
}

func (w WebError) Error() string {
	return w.err
}

func getValueFromServer(request string, name string) (string, error) {
	resp, err := http.Get(request)
	if err != nil {
		fmt.Printf("Sensor:%s GET error:%s\n", name, err.Error())
		return "", WebError{"Unable send GET request"}
	}

	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Printf("Relay:%s Read error:%s\n", name, err2.Error())
		return "", WebError{"getValueFromServer Unable resp.Body"}
	}

	defer resp.Body.Close()

	sBody := string(body)

	return sBody, nil

}
func prettyJSON(input string) (string, error) {
	var raw interface{}
	if err := json.Unmarshal([]byte(input), &raw); err != nil {
		return "", err
	}
	pretty, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return "", err
	}
	return string(pretty), nil
}

func jsonWrapper() js.Func {
	jsonfunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return "Invalid no of arguments passed"
		}
		jsDoc := js.Global().Get("document")
		if !jsDoc.Truthy() {
			return "Unable to get document object"
		}
		jsonOuputTextArea := jsDoc.Call("getElementById", "jsonoutput")
		if !jsonOuputTextArea.Truthy() {
			return "Unable to get output text area"
		}
		inputJSON := args[0].String()
		fmt.Printf("input %s\n", inputJSON)
		pretty, err := prettyJSON(inputJSON)
		if err != nil {
			errStr := fmt.Sprintf("unable to parse JSON. Error %s occurred\n", err)
			return errStr
		}
		jsonOuputTextArea.Set("value", pretty)
		return nil
	})

	return jsonfunc
}

func postActor(name string, action string) error {
	vals := url.Values{"Name": {name}, "Action": {action}}
	s := fmt.Sprintf("http://127.0.0.1:8090/setactor/%s/%s", vals["Name"][0], vals["Action"][0])

	jsDoc := js.Global().Get("document")
	if !jsDoc.Truthy() {
		return WebError{"Unable to get document object"}
	}

	actor := jsDoc.Call("getElementById", name)
	if !actor.Truthy() {
		return WebError{"Unable to find relay id " + name}
	}

	go func() {
		resp, err := http.PostForm(s, vals)
		if err != nil {
			fmt.Printf("Relay:%s POST error:%s\n", vals["Name"][0], err.Error())
			return
		}

		body, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			fmt.Printf("Relay:%s Read error:%s\n", vals["Name"][0], err2.Error())
		}

		defer resp.Body.Close()

		sBody := string(body)
		if sBody == "ON" {
			fmt.Printf("Status=%s\n", sBody)
			actor.Call("setAttribute", "style", "background:red;")
			if !actor.Truthy() {
				fmt.Printf("unable to set style background-color %s\n", name)
			}
		} else if sBody == "OFF" {
			fmt.Printf("Status=%s\n", sBody)
			actor.Call("setAttribute", "style", "background:black;")
			if !actor.Truthy() {
				fmt.Printf("unable to set style background-color %s\n", name)
			}
		} else {
			fmt.Printf("Unknown response = %s\n", sBody)
		}

	}()

	return nil

}
func postActorWapper() js.Func {
	actorFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 2 {
			return "Invalid no of arguments passed"
		}
		inputActorName := args[0].String()
		inputActorAction := args[1].String()
		fmt.Printf("input %s\n", inputActorName)
		err := postActor(inputActorName, inputActorAction)
		if err != nil {
			errStr := fmt.Sprintf("unable to Post  Actor Value. Error %s occurred\n", err)
			return errStr
		}
		return nil
	})

	return actorFunc
}

func onSetpointUpdate(name string) error {

	//fmt.Printf("onSetpointUpdate %s\n", name)
	request := fmt.Sprintf("http://127.0.0.1:8090/getsetpoint/%s", name)

	jsDoc := js.Global().Get("document")
	if !jsDoc.Truthy() {
		return WebError{"Unable to get document object"}
	}

	sensor := jsDoc.Call("getElementById", name)
	if !sensor.Truthy() {
		return WebError{"Unable to find sensor id " + name}
	}

	go func() {
		value, err := getValueFromServer(request, name)
		if err != nil {
			fmt.Printf("Unable to get document object")
			return
		}
		sensor.Set("innerText", value)

	}()

	return nil

}

func onSensorUpdate(name string) error {

	//fmt.Printf("onSensorUpdate %s\n", name)
	request := fmt.Sprintf("http://127.0.0.1:8090/getsensor/%s", name)

	jsDoc := js.Global().Get("document")
	if !jsDoc.Truthy() {
		return WebError{"Unable to get document object"}
	}

	sensor := jsDoc.Call("getElementById", name)
	if !sensor.Truthy() {
		return WebError{"Unable to find sensor id " + name}
	}

	go func() {
		value, err := getValueFromServer(request, name)
		if err != nil {
			fmt.Printf("Unable to get document object")
			return
		}
		sensor.Set("innerText", value)
	}()

	return nil

}

func onActorUpdate(name string) error {

	//fmt.Printf("onActorUpdate %s\n", name)
	s := fmt.Sprintf("http://127.0.0.1:8090/getactor/%s", name)

	jsDoc := js.Global().Get("document")
	if !jsDoc.Truthy() {
		return WebError{"Unable to get document object"}
	}

	actor := jsDoc.Call("getElementById", name)
	if !actor.Truthy() {
		fmt.Printf("getElementById::%s ERROR \n", name)
		return WebError{"relay:" + name + " getElementById error\n"}
	}

	go func() {
		resp, err := http.Get(s)
		if err != nil {
			fmt.Printf("relay:%s GET error:%s\n", name, err.Error())
			return
		}

		body, err2 := ioutil.ReadAll(resp.Body)
		if err2 != nil {
			fmt.Printf("Relay:%s Read error:%s\n", name, err2.Error())
			return
		}

		defer resp.Body.Close()

		sBody := string(body)
		actor.Set("innerText", sBody)

		if sBody == "ON" {
			//fmt.Printf("Status=%s\n", sBody)
			actor.Call("setAttribute", "style", "background:red;")
			if !actor.Truthy() {
				fmt.Printf("unable to set style background-color %s\n", name)
			}
		} else if sBody == "OFF" {
			//fmt.Printf("Status=%s\n", sBody)
			actor.Call("setAttribute", "style", "background:black;")
			if !actor.Truthy() {
				fmt.Printf("unable to set style background-color %s\n", name)
			}
		} else {
			fmt.Printf("'%s' Can't determine Actor status. Recieved: %s\n", name, sBody)
			actor.Call("setAttribute", "style", "background:grey;")
		}

	}()

	return nil

}

func main() {
	fmt.Println("Go Web Assembly")
	js.Global().Set("formatJSON", jsonWrapper())
	js.Global().Set("UpdateRelayValue", postActorWapper())
	//js.Global().Set("postSensorUpdate", postSensorUpdateWapper())
	sensors := []string{"Temp Sensor 1", "Temp Sensor 2", "Temp Sensor 3"}
	setpoints := []string{"Dummy Equipment 1"}
	actors := []string{"Relay 1", "Relay 2", "SSR 1"}

	//<-make(chan bool)
	t := time.NewTicker(5000 * time.Millisecond)

	for true {
		<-t.C
		for _, sensor := range sensors {
			err := onSensorUpdate(sensor)
			if err != nil {
				fmt.Printf("sensor read error: %s\n", err.Error())
			}
		}
		for _, setpoint := range setpoints {
			err := onSetpointUpdate(setpoint)
			if err != nil {
				fmt.Printf("Setpoint read error: %s\n", err.Error())
			}
		}
		for _, actor := range actors {
			err := onActorUpdate(actor)
			if err != nil {
				fmt.Printf("actor read error: %s\n", err.Error())
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}
