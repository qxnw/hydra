

header={}
header["Content-Type"]="text/plain"

function main(p)
    local r={header={}}
    r.header["Location"]="/order/request"  
    return r
end