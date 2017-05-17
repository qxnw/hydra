
header={}
header["Charset"]="gbk"

function main(p)
    p["header"]={}
    p["header"]["Content-type"]="text/plain"
    p["header"]={id=100,name="colin"}
    return p
end