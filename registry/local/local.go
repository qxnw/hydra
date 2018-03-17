package local

import (
	"os"
	"strings"
)
type local struct{

}
func(l *local)Exists(path string) (bool, error){
    rpath:=path
	if strings.HasPrefix(path,"."){
		rpath="."+path
	}
	_,err:=os.Stat(path)
	if err==nil{
		return true,nil
	}
	if os.Exists(err){
		return false,nil
	}
	return false,err
}
