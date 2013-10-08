package parser

import (
    "github.com/bmizerany/assert"
    "testing"
)

func TestDefaultParserParseLine(t *testing.T) {
    line := `us,1381118458069,{"cheater":10301051,"type":"helpFriendsRewardAction","world_id":"100001823535095","user":"100001823535095","_log_info":{"uid":10301051,"script_id":3183040714,"serial":3,"host":"10.255.8.189","ip":"79.215.100.157"}}`
    p := new(DefaultParser)
    area, ts, data := p.ParseLine(line)
    var (
        exptectedTs  = uint64(1381118458069 / 1000)
        extectedArea = "us"
    )

    if area != extectedArea {
        t.Error("area: expected", extectedArea, "got", area)
    }
    assert.Equal(t, exptectedTs, ts)

    typ, err := data.Get("type").String()
    if err != nil {
        t.Error(err)
    }
    assert.Equal(t, "helpFriendsRewardAction", typ)
    var expectedScriptId int64 = 3183040714
    scriptId, err := data.Get("_log_info").Get("script_id").Int64()
    if err != nil {
        t.Error(err)
    }
    assert.Equal(t, expectedScriptId, scriptId)
}

func TestExtractLogInfo(t *testing.T) {
    line := `us,1381118458069,{"cheater":10301051,"type":"helpFriendsRewardAction","world_id":"100001823535095","user":"100001823535095","_log_info":{"uid":10301051,"script_id":3183040714,"serial":3,"host":"10.255.8.189","ip":"79.215.100.157"}}`
    p := new(DefaultParser)
    _, _, data := p.ParseLine(line)
    info := extractLogInfo(data)
    assert.Equal(t, info.host, "10.255.8.189")
    assert.Equal(t, info.ip, "79.215.100.157")
    var (
        expectedScriptId int64 = 3183040714
        expectedUid      int64 = 10301051
    )
    assert.Equal(t, expectedScriptId, info.scriptId)
    assert.Equal(t, info.serial, 3)
    assert.Equal(t, expectedUid, info.uid)
}
