package statusCode

const SmsCodeSendTimesOut = 913       //短信发送次数超过限制
const SmsCheckCodeErrorTimesOut = 901 //验证码验证错误次数超过啦限制
const ImgCodeCheckError = 902         //图片验证码填写错误
const MobileNoError = 903             //电话号码错误
const SmsCodeCheckError = 904         //短信验证码填写错误
const PromotionNotThisUser = 905      //该用户不能参加这个活动

const ActivityNotStart = 906        //活动还没有开始
const ActivityIsEnd = 907           //活动已经结束
const ActivityIsPause = 908         //活动被暂停
const CustomerJoinedActivity = 909  //用户已经参加过这个活动
const CustomerNotJoinActivity = 910 //用户不能参加活动

const CouponPlanStatusError = 911   //优惠卷方案状态错误
const CouponNumberIsNotEnough = 912 //优惠卷可用数量不足
const CouponAmountIsNotEnough = 914 //优惠卷可领取金额不足

const CouponStatusHasUsed = 921 //1.已使用
const CouponStatusExpire = 922  //2.已过期

const LoginAccountOrPasswordEmpty = 950 //用户名或密码为空
const LoginAccountOrPasswordError = 951 //用户名或密码错误
const LoginAccountLocked = 952          //帐户被锁定
const LoginAccountDisabled = 953        //帐户被禁用
const LoginAccountOverMax = 954         //登录次数最大次数，帐户被锁定
const LoginAccountDbError = 955         //数据库错误
const LoginAccountLost = 956            //帐户不存在
const LoginAccountLogged = 957          //帐户已登录

const ConsumeStationRefuse = 990 //加油站未参与优惠券活动
const ConsumeStationExpire = 991 //优惠券已过期
const ConsumeStationUsed = 992   //优惠券已使用
const ConsumeStationUnrule = 993 //优惠券不符合使用规则
