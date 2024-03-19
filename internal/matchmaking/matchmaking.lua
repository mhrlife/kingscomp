-- Check queue size and match users
local function matchUsers(queueKey, pubSubChannel, minUsers, minScore, lobbyId, userId, userScore)
    -- Get users within the time range
    local users = redis.call('ZRANGEBYSCORE', queueKey, minScore, '+inf', 'LIMIT', 0, minUsers)
    if #users >= minUsers then
        -- convert strings to ints
        for i, v in ipairs(users) do
            users[i] = tonumber(v)
        end

        table.insert(users, userId)
        -- Remove these users from the sorted set
        redis.call('ZREM', queueKey, unpack(users))
        ---- Create a new lobby
        --local lobby = {
        --    id = lobbyId,
        --    participants = users,
        --    created_at = userScore,
        --    state = 'created'
        --}
        --local lobbyJson = cjson.encode(lobby)
        --redis.call('JSON.SET', 'lobby:' .. lobbyId, '.', lobbyJson)
        ---- Notify the matched users via Pub/Sub channel
        --for i, v in ipairs(users) do
        --    if v ~= userId then
        --        local listKey = 'matchmaking:' .. v
        --        redis.call('RPUSH', listKey, lobbyId)
        --        redis.call('EXPIRE', listKey, 120)
        --    end
        --end
        return { true, lobbyId, users } -- Matching succeeded
    else
        -- Add the current user to the queue since not enough users are present
        redis.call('ZADD', queueKey, userScore, userId)
        return { false } -- User added to queue, not enough users for matching
    end
    return { false } -- Not enough users for matching
end

-- Keys and arguments
local queueKey = KEYS[1]
local pubSubChannel = KEYS[2]

local minUsers = tonumber(ARGV[1])
local minScore = tonumber(ARGV[2])
local lobbyId = ARGV[3]
local userId = tonumber(ARGV[4])
local userScore = tonumber(ARGV[5])

-- Call the function and return its result
return matchUsers(queueKey, pubSubChannel, minUsers, minScore, lobbyId, userId, userScore)