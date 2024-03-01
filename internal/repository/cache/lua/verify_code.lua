local key = KEYS[1]
-- key使用次数，验证次数
local cntKey = key..":cnt"
-- 用户输入的验证码
local expectedCode=ARGV[1]

local cnt = tonumber(redis.call("get",cntKey))
local code = redis.call("get",key)


if cnt == nil or cnt<=0 then
    -- 说明验证次数耗尽了
    return -1
end

-- 比较验证码
if expectedCode==code then
    redis.call("set",cntKey,0)
    return 0
else
    -- 不相等，用户输出错误，减去验证次数
    redis.call("decr",key)
    return -2
end