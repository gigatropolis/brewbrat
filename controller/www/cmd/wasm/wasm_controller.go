package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"syscall/js"
)

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

		sBody := string(body)
		if sBody == "ON" || sBody == "OFF" {
			fmt.Printf("Status=%s\n", sBody)
		} else {
			fmt.Printf("Unknown response = %s\n", sBody)
		}

		resp.Body.Close()
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

func main() {
	fmt.Println("Go Web Assembly")
	js.Global().Set("formatJSON", jsonWrapper())
	js.Global().Set("UpdateRelayValue", postActorWapper())
	<-make(chan bool)
}
