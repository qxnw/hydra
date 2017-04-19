
_header={}
_header["Charset"]="gbk"

function main(p)
    p["_header"]={}
    p["_header"]["Content-type"]="text/plain"
    p["data"]={id=100,name="colin"}
    return p
end