package cacheKey

type CKV struct {
	Key     string
	Expires int
}
var (
	UserLogin= &CKV{"coupon-system:customer:login:@customerId", 6000}	
)
