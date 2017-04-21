

_header={}
_header["Content-Type"]="text/plain"

function main(p)
    local r={_header={}}
    r._header["Location"]="/order/request"  
    return r
end