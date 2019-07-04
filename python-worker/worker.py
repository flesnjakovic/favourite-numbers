import asyncio
import redis
import collections

redis = redis.Redis(host='redis', port=6379, db=0)

def publishUsers():
    users = redis.hgetall("users")
    sortedUsers = collections.OrderedDict(sorted(users.items()))

    output_msg = ""
    for k, v in sortedUsers.items():
        output_msg += "\n" + k.decode('utf-8') + " " + v.decode('utf-8')

    redis.publish("user_list", output_msg)

def setFavouriteNumber(msg):
    redis.hset("users", msg[0], int(msg[1]))

def receiveWork(pubsub):
    for msg in pubsub.listen():
        if msg["type"] == "message":
            data = msg["data"].decode("utf-8").split()
            if data[0] == "list":
                publishUsers()
                continue
            
            setFavouriteNumber(data)
            publishUsers()

if __name__ == '__main__':
    pubsub = redis.pubsub()
    pubsub.subscribe("work")
    asyncio.async(receiveWork(pubsub))