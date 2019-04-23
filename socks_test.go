package Socks5

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
)

func TestBaidu(t *testing.T) {
	wait := sync.WaitGroup{}
	for i := 0; i < 20000; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			client := http.DefaultClient
			response, err := client.Get("http://www.baidu.com")
			if err != nil {
				return
			}
			bytes, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return
			}
			fmt.Println(string(bytes[:10]))
		}()
	}
	wait.Wait()
}
