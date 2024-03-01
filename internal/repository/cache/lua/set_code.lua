local key = KEYS[1]
-- key使用次数，验证次数
local cntKey = key..":cnt"
-- 准备存储的验证码
local val=ARGV[1]
-- 使用ttl命令查看Key的剩余生存时间
local ttl= tonumber(redis.call("ttl",key))

if ttl == -1 then
    -- key存在，但是没有过期时间
    return -2
elseif ttl == -2 or ttl <540 then
    -- key不存在或者剩余生存时间小于9分钟可以发送验证码，验证码1分钟发送一条
    redis.call("set",key,val)
    redis.call("expire",key,600)
    redis.call("set",cntKey,3)
    redis.call("expire",cntKey,600)
else
    -- 发送太频繁
    return -1
end