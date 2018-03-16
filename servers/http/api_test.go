package http

import(
	"testing"
	"reflect"
	"fmt"
	"context"
	"time"
)

func Expect(t *testing.T, a interface{}, b interface{}) {
    if a != b {
        t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
    }
}

func Refute(t *testing.T, a interface{}, b interface{}) {
    if a == b {
        t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
    }
}

func Test_getAddress_1(t *testing.T){
	apiServer:=&ApiServer{}

	if addr1:= apiServer.getAddress();addr1!="0.0.0.0:8000"{
		a:=addr1
		b:="0.0.0.0:8000"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}

	if addr1:= apiServer.getAddress(nil);addr1!="0.0.0.0:8000"{
		a:=addr1
		b:="0.0.0.0:8000"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}

	if addr1:= apiServer.getAddress("");addr1!="0.0.0.0:8000"{
		a:=addr1
		b:="0.0.0.0:8000"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
	
	if addr1:= apiServer.getAddress(-300);addr1!="0.0.0.0:8000"{
		a:=addr1
		b:="0.0.0.0:8000"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}

	if addr1:= apiServer.getAddress(0);addr1!="0.0.0.0:8000"{
		a:=addr1
		b:="0.0.0.0:8000"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}

	if addr1:= apiServer.getAddress(3.14);addr1!="0.0.0.0:8000"{
		a:=addr1
		b:="0.0.0.0:8000"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}

	if addr1:= apiServer.getAddress(300);addr1!="0.0.0.0:300"{
		a:=addr1
		b:="0.0.0.0:300"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}

	if addr1:= apiServer.getAddress("abc");addr1!="abc:8000"{
		a:=addr1
		b:="abc:8000"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}

	if addr1:= apiServer.getAddress("192.168.0.106");addr1!="192.168.0.106:8000"{
		a:=addr1
		b:="192.168.0.106:8000"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}

	if addr1:= apiServer.getAddress("192.168.0.106","abc");addr1!="192.168.0.106:8000"{
		a:=addr1
		b:="192.168.0.106:8000"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}

	if addr1:= apiServer.getAddress("abc","123");addr1!="abc:123"{
		a:=addr1
		b:="abc:123"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}

	if addr1:= apiServer.getAddress("192.168.0.154","123");addr1!="192.168.0.154:123"{
		a:=addr1
		b:="192.168.0.154:123"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}

	if addr1:= apiServer.getAddress("192.168.0.154",123);addr1!="192.168.0.154:123"{
		a:=addr1
		b:="192.168.0.154:123"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}

	if addr1:= apiServer.getAddress("192.168.0.154",123,"cdsss");addr1!="192.168.0.154:123"{
		a:=addr1
		b:="192.168.0.154:123"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}

	if addr1:= apiServer.getAddress(456,123,"cdsss");addr1!="0.0.0.0:123"{
		a:=addr1
		b:="0.0.0.0:123"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

type testEngine1 struct{

}
func (t *testEngine1)ListenAndServe() error{
	return  fmt.Errorf("%v", "监听线程出错")
}
func (t *testEngine1)ListenAndServeTLS(certFile, keyFile string) error{
	return  fmt.Errorf("%v", "监听线程出错")
}
func (t *testEngine1)Shutdown(ctx context.Context) error{
	return nil
}

func Test_Run_1(t *testing.T){
	apiServer:=&ApiServer{}
	apiServer.engine=&testEngine1{}
	if err:= apiServer.Run("");err.Error()!="监听线程出错"{
		a:=err.Error()
		b:="监听线程出错"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
	if status:= apiServer.GetStatus();status!="stop"{
		a:=status
		b:="stop"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

type testEngine2 struct{

}
func (t *testEngine2)ListenAndServe() error{
	for i:=0;i<1000;i++{
		time.Sleep(1000)
	}
	return  nil
}
func (t *testEngine2)ListenAndServeTLS(certFile, keyFile string) error{
	for i:=0;i<1000;i++{
		time.Sleep(1000)
	}
	return  nil
}
func (t *testEngine2)Shutdown(ctx context.Context) error{
	return nil
}

func Test_Run_2(t *testing.T){
	apiServer:=&ApiServer{}
	apiServer.engine=&testEngine2{}
	if err:= apiServer.Run("");err!=nil{
		a:=err.Error()
		b:="nil"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
	if status:= apiServer.GetStatus();status!="running"{
		a:=status
		b:="running"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}


func Test_Run_3(t *testing.T){
	apiServer:=&ApiServer{}
	apiServer.engine=&testEngine1{}
	if err:= apiServer.RunTLS("","","");err.Error()!="监听线程出错"{
		a:=err.Error()
		b:="监听线程出错"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
	if status:= apiServer.GetStatus();status!="stop"{
		a:=status
		b:="stop"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func Test_Run_4(t *testing.T){
	apiServer:=&ApiServer{}
	apiServer.engine=&testEngine2{}
	if err:= apiServer.RunTLS("","","");err!=nil{
		a:=err.Error()
		b:="nil"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
	if status:= apiServer.GetStatus();status!="running"{
		a:=status
		b:="running"
		t.Errorf("Expected %s (type %s) - Got %s (type %s)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}