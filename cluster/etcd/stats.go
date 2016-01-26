package etcd

/*
/v2/stats/self
{
    "name": "default",
    "id": "ce2a822cea30bfca",
    "state": "StateLeader",
    "startTime": "2016-01-26T14:51:30.694261426+08:00",
    "leaderInfo": {
        "leader": "ce2a822cea30bfca",
        "uptime": "1h22m35.435654084s",
        "startTime": "2016-01-26T14:51:31.095511161+08:00"
    },
    "recvAppendRequestCnt": 0,
    "sendAppendRequestCnt": 0
}
*/
type StatSelf struct {
}

type StatLeader struct {
}

/*
/v2/stats/store
{
    "compareAndSwapFail": 0,
    "compareAndSwapSuccess": 0,
    "createFail": 0,
    "createSuccess": 2,
    "deleteFail": 0,
    "deleteSuccess": 0,
    "expireCount": 0,
    "getsFail": 4,
    "getsSuccess": 75,
    "setsFail": 2,
    "setsSuccess": 4,
    "updateFail": 0,
    "updateSuccess": 0,
    "watchers": 0
}
*/
type StatStore struct {
}
