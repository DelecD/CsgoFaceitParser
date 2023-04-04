package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/gogo/protobuf/test/issue312/events"
	demoinfocs "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/events"
)

//Настройки
type Settings struct {
	CompetitionIds        []string
	PagesRange            []int
	PageSize              int
	Verbose               bool
	MatchesParallels      int
	WaitIf403ErrorSecs    int
	RepeatAfterFinish     bool
	DelayBeforeRepeatSecs int
	SaveDemoFiles         bool
	FaceitApiKey          string
	MysqlUser             string
	MysqlPassword         string
	MysqlHost             string
	MysqlPort             string
	MysqlDbName           string
}

var settings Settings = Settings{
	CompetitionIds:        []string{"42e160fc-2651-4fa5-9a9b-829199e27adb"},
	PagesRange:            []int{1, 10},
	PageSize:              100,
	Verbose:               true,
	MatchesParallels:      8,
	WaitIf403ErrorSecs:    10,
	RepeatAfterFinish:     true,
	DelayBeforeRepeatSecs: 300,
	FaceitApiKey:          "79f42c0d-fb4f-47b6-8589-9373881c5a71",
	SaveDemoFiles:         true,
	MysqlUser:             "root",
	MysqlPassword:         "root",
	MysqlHost:             "localhost",
	MysqlPort:             "3306",
	MysqlDbName:           "testevents",
}

//Базовая структура для событий, чтобы разные события можно было хранить в одном массиве
//Если бы в go было нормальное ООП, реализовал бы через базовый класс, но...
type IFace struct {
	typeId int
	value  interface{}
}

type MatchInfo struct {
	MatchId    string
	Team1      string
	Team2      string
	DemoPath   string
	Map        string
	MapId      int
	StartedAt  string
	FinishedAt string
	ParsedAt   string
}

//Реплика таблиц info_maps, info_weapons из БД, чтобы лишний раз не дергать ее
type WeaponDbRecord struct {
	Id   int
	Name string
}
type MapDbRecord struct {
	Id   int
	Name string
}

var weaponsDbTable []WeaponDbRecord = make([]WeaponDbRecord, 0)
var mapsDbTable []MapDbRecord = make([]MapDbRecord, 0)

//События
type KillEvent struct {
	killerSteamId      uint64
	killerName         string
	killerHp           int8
	killerFaceitLevel  int
	killerKilledBefore int8
	victimSteamId      uint64
	victimName         string
	victimFaceitLevel  int
	aliveCT            int8
	aliveT             int8
	weaponName         string
	weaponId           int
	flags              int32
	timeSecond         int32
	timeFrame          int
	timeRound          int16
}

type BombPlantedEvent struct {
	playerSteamId     uint64
	playerName        string
	playerHp          int8
	playerFaceitLevel int
	aliveCT           int8
	aliveT            int8
	flags             int32
	timeSecond        int32
	timeFrame         int
	timeRound         int16
}

type BombDefusedEvent struct {
	playerSteamId     uint64
	playerName        string
	playerHp          int8
	playerFaceitLevel int
	aliveCT           int8
	aliveT            int8
	flags             int32
	timeSecond        int32
	timeFrame         int
	timeRound         int16
}

//Флаги событий
var FLAG_HEADSHOT int32 = 1
var FLAG_THROUGH_SMOKE int32 = 2
var FLAG_BLINDED int32 = 4
var FLAG_NO_SCOPE int32 = 8
var FLAG_BOT_KILLER int32 = 16
var FLAG_PENETRATION int32 = 32
var FLAG_ON_AIR int32 = 64
var FLAG_MOLOTOV int32 = 128

//Типы событий
var EVENT_KILLEVENT = 1
var EVENT_BOMBPLANT = 2
var EVENT_BOMBDEFUSE = 3

var db *sql.DB
var matchInfos []MatchInfo = make([]MatchInfo, 0)
var err error

func main() {
	settingsLoad()
	checkError(err)

	if _, err := os.Stat("demos"); os.IsNotExist(err) {
		err = os.Mkdir("demos", fs.ModeDir)
		if err != nil {
			panic(err)
		}
	}

	sqlConParam := settings.MysqlUser + ":" +
		settings.MysqlPassword + "@tcp(" +
		settings.MysqlHost + ":" +
		settings.MysqlPort + ")/" +
		settings.MysqlDbName
	db, err = sql.Open("mysql", sqlConParam)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	resp, _ := db.Query("SELECT * FROM info_weapons")
	if resp != nil {
		defer resp.Close()
		for resp.Next() {
			var record = WeaponDbRecord{}
			resp.Scan(&record.Id, &record.Name)
			weaponsDbTable = append(weaponsDbTable, record)
		}
	}
	resp, _ = db.Query("SELECT * FROM info_maps")
	if resp != nil {
		defer resp.Close()
		for resp.Next() {
			var record = MapDbRecord{}
			resp.Scan(&record.Id, &record.Name)
			mapsDbTable = append(mapsDbTable, record)
		}
	}

	for {
		matchesLeft := 0

		if settings.Verbose {
			fmt.Println("Загружаем список демок")
		}

		for _, competitionId := range settings.CompetitionIds {
			for page := settings.PagesRange[0]; page < settings.PagesRange[1]+1; page++ {
				url := "https://api.faceit.com/match-history/v4/matches/competition?id=" +
					competitionId +
					"&page=" +
					strconv.Itoa(page) +
					"&size=" +
					strconv.Itoa(settings.PageSize) +
					"&type=matchmaking"
				jsonRes := parseJson(getUrl(url))
				matches := jsonRes["payload"].([]interface{})
				for _, matchJson := range matches {
					matchData := matchJson.(map[string]interface{})
					if matchData["state"].(string) == "finished" {
						matchId := matchData["matchId"].(string)
						team1 := matchData["faction1Nickname"].(string)
						team2 := matchData["faction2Nickname"].(string)
						startedAt := matchData["startedAt"].(string)
						finishedAt := matchData["finishedAt"].(string)
						startedAt = strings.Replace(startedAt, "T", " ", 1)
						startedAt = strings.Replace(startedAt, "Z", "", 1)
						finishedAt = strings.Replace(startedAt, "T", " ", 1)
						finishedAt = strings.Replace(startedAt, "Z", "", 1)

						matchData := MatchInfo{
							MatchId:    matchId,
							Team1:      team1,
							Team2:      team2,
							StartedAt:  startedAt,
							FinishedAt: finishedAt,
						}
						matchInfos = append(matchInfos, matchData)
						matchesLeft++
					}
				}
				if settings.Verbose {
					fmt.Println("Загружена страница", page, ", competitionId: ", competitionId)
				}
			}
		}
		if settings.Verbose {
			fmt.Println("Список составлен, начинаем...")
		}

		matchesAwaiterCh := make(chan bool)
		matchesHandled := 0

		for _, matchInfo := range matchInfos {
			matchesHandled++
			go handleMatch(matchInfo, matchesAwaiterCh)
			if matchesHandled == settings.MatchesParallels {
				<-matchesAwaiterCh
				matchesHandled--
				matchesLeft--
				if settings.Verbose {
					fmt.Println("Осталось ", matchesLeft, " из ", len(matchInfos))
				}
			}
		}

		for matchesHandled > 0 {
			<-matchesAwaiterCh
			matchesHandled--
			matchesLeft--
			if settings.Verbose {
				fmt.Println("Осталось ", matchesLeft, " из ", len(matchInfos))
			}
		}

		if !settings.RepeatAfterFinish {
			break
		}
		if settings.Verbose {
			fmt.Println("Парсинг закончен. Через ", settings.DelayBeforeRepeatSecs, " секунд начнем заново...")
		}
		time.Sleep(time.Duration(settings.DelayBeforeRepeatSecs) * time.Second)
	}

	var input string
	fmt.Print("\n\nПарсинг закончен\nНажмите любую клавишу для выхода")
	fmt.Scanln(&input)
}

//Обработка матча
func handleMatch(match MatchInfo, matchesAwaiterCh chan bool) {
	//Получение информации о матче с сервера Faceit
	urlParams := map[string]string{
		"Authorization": "Bearer " + settings.FaceitApiKey,
	}
	matchPage := urlRequest("https://open.faceit.com/data/v4/matches/"+match.MatchId, "GET", urlParams)
	jsonMatchPage := parseJson(matchPage)

	demosPaths := jsonMatchPage["demo_url"].([]interface{})

	//Получение FACEIT levels
	faceitLevels := make(map[uint64]int)
	jsonTeams := jsonMatchPage["teams"].(map[string]interface{})
	jsonTeam1 := jsonTeams["faction1"].(map[string]interface{})
	jsonTeam2 := jsonTeams["faction2"].(map[string]interface{})
	for _, ply := range jsonTeam1["roster"].([]interface{}) {
		jsonPlayer := ply.(map[string]interface{})
		plySteamId, _ := strconv.ParseInt(jsonPlayer["game_player_id"].(string), 10, 64)
		plyFaceitLvl := jsonPlayer["game_skill_level"].(float64)
		faceitLevels[uint64(plySteamId)] = int(plyFaceitLvl)
	}
	for _, ply := range jsonTeam2["roster"].([]interface{}) {
		jsonPlayer := ply.(map[string]interface{})
		plySteamId, _ := strconv.ParseInt(jsonPlayer["game_player_id"].(string), 10, 64)
		plyFaceitLvl := jsonPlayer["game_skill_level"].(float64)
		faceitLevels[uint64(plySteamId)] = int(plyFaceitLvl)
	}

	doneChannel := make(chan bool)
	for i := 0; i < len(demosPaths); i++ {
		demoPath := demosPaths[i]
		go handleDemo(match, faceitLevels, demoPath.(string), doneChannel)
		result := <-doneChannel
		if !result {
			i--
			time.Sleep(time.Duration(settings.WaitIf403ErrorSecs) * time.Second)
		}
	}
	matchesAwaiterCh <- true
}

func handleDemo(match MatchInfo, faceitLevels map[uint64]int, demoPath string, doneChannel chan bool) {
	//Проверка на существование демки в базе
	rows, err := db.Query("SELECT 1 FROM demos WHERE faceitId = '" + match.MatchId + "'")
	if err != nil {
		fmt.Println(err.Error(), demoPath)
		doneChannel <- false
		return
	}
	//Если демка есть в базе
	if rows != nil {
		if rows.Next() {
			if settings.Verbose {
				fmt.Print("Пропускаем обработанное демо ", match.MatchId)
			}
			doneChannel <- true
			return
		}
	}

	//Выгружаем файл демки
	urlParams := map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:96.0) Gecko/20100101 Firefox/96.0",
	}
	pathParts := strings.Split(demoPath, "/")
	filename := pathParts[len(pathParts)-1]
	err = downloadFile("demos"+string(os.PathSeparator)+filename, demoPath, "GET", urlParams)
	if err != nil {
		fmt.Println(err.Error(), demoPath)
		doneChannel <- false
		return
	}
	eventList := analyzeGzip(&match, faceitLevels, "demos"+string(os.PathSeparator)+filename)
	if !settings.SaveDemoFiles {
		os.Remove("demos" + string(os.PathSeparator) + filename)
	}

	//Заносим в таблицу demos данные о спарсенной демке
	result, err := db.Exec("INSERT INTO demos VALUES (NULL, ?, ?, ?, ?, ?, ?, ?, ?)",
		match.MatchId, match.Team1, match.Team2, match.MapId,
		match.ParsedAt, match.StartedAt, match.FinishedAt,
		match.DemoPath)
	if err != nil {
		fmt.Errorf(err.Error())
		doneChannel <- false
		return
	}

	//Заносим в БД все собранные парсером события
	newMatchIdInDb, _ := result.LastInsertId()
	for _, eventData := range eventList {
		eventType := eventData.typeId
		switch eventType {
		case EVENT_KILLEVENT:
			ev := eventData.value.(*KillEvent)
			_, err = db.Exec("INSERT INTO events_kill VALUES (NULL,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
				newMatchIdInDb,
				ev.killerSteamId, ev.killerName, ev.killerHp, ev.killerFaceitLevel,
				ev.victimSteamId, ev.victimName, ev.victimFaceitLevel,
				ev.weaponId,
				ev.timeRound, ev.timeSecond, ev.timeFrame,
				ev.aliveCT, ev.aliveT,
				ev.flags)
		case EVENT_BOMBPLANT:
			ev := eventData.value.(*BombPlantedEvent)
			_, err = db.Exec("INSERT INTO events_bomb VALUES (NULL,?,?,?,?,?,?,?,?,?,?,?,?)",
				newMatchIdInDb, 0,
				ev.playerSteamId, ev.playerName, ev.playerFaceitLevel, ev.playerHp,
				ev.aliveCT, ev.aliveT,
				ev.flags,
				ev.timeRound, ev.timeSecond, ev.timeFrame)
		case EVENT_BOMBDEFUSE:
			ev := eventData.value.(*BombDefusedEvent)
			_, err = db.Exec("INSERT INTO events_bomb VALUES (NULL,?,?,?,?,?,?,?,?,?,?,?,?)",
				newMatchIdInDb, 1,
				ev.playerSteamId, ev.playerName, ev.playerFaceitLevel, ev.playerHp,
				ev.aliveCT, ev.aliveT,
				ev.flags,
				ev.timeRound, ev.timeSecond, ev.timeFrame)
		}
		if err != nil {
			fmt.Errorf(err.Error())
			doneChannel <- false
			return
		}
	}
	doneChannel <- true
}

//Извлечение демки из архива *.gz, запуск процедуры parseDemoFile(), удаление извлеченной
//демки (не архива)
func analyzeGzip(match *MatchInfo, faceitLevels map[uint64]int, zippath string) []IFace {
	f, err := os.Open(zippath)
	if err != nil {
		print(err)
		return nil
	}
	defer f.Close()

	archive, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}

	demofileName := archive.Name
	unpackedDemoPath := "demos" + string(os.PathSeparator) + demofileName
	file_write, err := os.Create(unpackedDemoPath)
	if _, err = io.Copy(file_write, archive); err != nil {
		panic(err)
	}

	mapName, eventList := parseDemoFile(unpackedDemoPath, faceitLevels)
	os.Remove(unpackedDemoPath)
	match.DemoPath = zippath
	match.ParsedAt = time.Now().Format("2006-01-02 15:04:05")

	mapId := -1
	for _, v := range mapsDbTable {
		if v.Name == mapName {
			mapId = v.Id
		}
	}
	if mapId == -1 {
		db.Exec("INSERT INTO info_maps VALUES(NULL, '" + mapName + "')")
		mapId = len(mapsDbTable)
		row := db.QueryRow("SELECT id FROM info_maps WHERE name = '" + mapName + "'")
		row.Scan(&mapId)
		mapsDbTable = append(mapsDbTable, MapDbRecord{
			Id:   mapId,
			Name: mapName,
		})
	}
	match.Map = mapName
	match.MapId = mapId
	return eventList
}

//Парсинг демо-файла
func parseDemoFile(path string, faceitLevels map[uint64]int) (string, []IFace) {
	var eventList []IFace = make([]IFace, 0)

	if settings.Verbose {
		fmt.Println("Начинаем анализ " + path + "...")
	}

	f, err := os.Open(path)
	checkError(err)
	defer f.Close()

	p := demoinfocs.NewParser(f)
	defer p.Close()

	var round int = 0
	killsCount := make(map[uint64]int)

	mapName := ""
	p.RegisterEventHandler(func(events.DataTablesParsed) {
		mapName = p.Header().MapName
	})

	p.RegisterEventHandler(func(e events.Kill) {
		if p.GameState().IsWarmupPeriod() {
			return
		}
		if e.Killer == nil || e.Weapon == nil || e.Weapon.Entity == nil {
			return
		}
		if e.Weapon.Type == common.EqWorld {
			return
		}
		if p.GameState().TotalRoundsPlayed() > round {
			for k := range killsCount {
				killsCount[k] = 0
			}
			round += 1
		}

		killerSteamId := e.Killer.SteamID64
		victimSteamId := e.Victim.SteamID64
		throughSmoke := e.ThroughSmoke
		attackerBlind := e.AttackerBlind
		noScope := e.NoScope
		penetratedObjectsCount := e.PenetratedObjects
		headshot := e.IsHeadshot
		botKiller := e.Killer.IsBot
		onAir := e.Victim.IsAirborne()
		molotov := false

		aliveT, aliveCT := CalcAliveT_CT(p, victimSteamId)

		if e.Weapon.Type == common.EqMolotov || e.Weapon.Type == common.EqIncendiary {
			molotov = true
		}

		flags := 0
		if throughSmoke {
			flags |= int(FLAG_THROUGH_SMOKE)
		}
		if attackerBlind {
			flags |= int(FLAG_BLINDED)
		}
		if noScope {
			flags |= int(FLAG_NO_SCOPE)
		}
		if penetratedObjectsCount > 0 {
			flags |= int(FLAG_PENETRATION)
		}
		if headshot {
			flags |= int(FLAG_HEADSHOT)
		}
		if botKiller {
			flags |= int(FLAG_BOT_KILLER)
		}
		if onAir {
			flags |= int(FLAG_ON_AIR)
		}
		if molotov {
			flags |= int(FLAG_MOLOTOV)
		}

		if flags != 0 || killsCount[killerSteamId] > 0 {
			//Проверяем, есть ли имя оружия в таблице info_weapons, если нет - добавляем
			//в таблицу в БД и в ее реплику
			weaponId := -1
			weaponName := e.Weapon.String()
			for _, v := range weaponsDbTable {
				if v.Name == weaponName {
					weaponId = v.Id
				}
			}
			if weaponId == -1 {
				db.Exec("INSERT INTO info_weapons VALUES(NULL, '" + weaponName + "')")
				row := db.QueryRow("SELECT id FROM info_weapons WHERE name = '" + weaponName + "'")
				row.Scan(&weaponId)
				weaponsDbTable = append(weaponsDbTable, WeaponDbRecord{
					Id:   weaponId,
					Name: weaponName,
				})
			}

			event := KillEvent{
				killerSteamId:      killerSteamId,
				killerName:         e.Killer.Name,
				killerHp:           int8(e.Killer.Health()),
				killerFaceitLevel:  faceitLevels[killerSteamId],
				victimSteamId:      victimSteamId,
				victimName:         e.Victim.Name,
				victimFaceitLevel:  faceitLevels[victimSteamId],
				killerKilledBefore: int8(killsCount[killerSteamId]),
				aliveCT:            int8(aliveCT),
				aliveT:             int8(aliveT),
				weaponName:         weaponName,
				weaponId:           weaponId,
				flags:              int32(flags),
				timeSecond:         int32(p.CurrentTime().Seconds()),
				timeFrame:          p.GameState().IngameTick(),
				timeRound:          int16(p.GameState().TotalRoundsPlayed()) + 1,
			}

			eventList = append(eventList, IFace{
				typeId: EVENT_KILLEVENT,
				value:  &event,
			})
		}

		killsCount[killerSteamId] += 1
	})

	p.RegisterEventHandler(func(e events.BombPlanted) {
		aliveT, aliveCT := CalcAliveT_CT(p, 0)
		event := BombPlantedEvent{
			playerSteamId:     e.Player.SteamID64,
			playerName:        e.Player.Name,
			playerHp:          int8(e.Player.Health()),
			playerFaceitLevel: faceitLevels[e.Player.SteamID64],
			aliveCT:           int8(aliveCT),
			aliveT:            int8(aliveT),
			flags:             0,
			timeSecond:        int32(p.CurrentTime().Seconds()),
			timeFrame:         p.GameState().IngameTick(),
			timeRound:         int16(p.GameState().TotalRoundsPlayed()) + 1,
		}
		eventList = append(eventList, IFace{
			typeId: EVENT_BOMBPLANT,
			value:  &event,
		})
	})

	p.RegisterEventHandler(func(e events.BombDefused) {
		aliveT, aliveCT := CalcAliveT_CT(p, 0)
		event := BombDefusedEvent{
			playerSteamId:     e.Player.SteamID64,
			playerName:        e.Player.Name,
			playerHp:          int8(e.Player.Health()),
			playerFaceitLevel: faceitLevels[e.Player.SteamID64],
			aliveCT:           int8(aliveCT),
			aliveT:            int8(aliveT),
			flags:             0,
			timeSecond:        int32(p.CurrentTime().Seconds()),
			timeFrame:         p.GameState().IngameTick(),
			timeRound:         int16(p.GameState().TotalRoundsPlayed()) + 1,
		}
		eventList = append(eventList, IFace{
			typeId: EVENT_BOMBDEFUSE,
			value:  &event,
		})
	})

	err = p.ParseToEnd()
	if err != nil {
		fmt.Println(err)
	}

	return mapName, eventList
}

//Подсчет живых терров и контров
func CalcAliveT_CT(p demoinfocs.Parser, victimSteamId uint64) (int, int) {
	var aliveT int = 0
	var aliveCT int = 0
	for _, ply := range p.GameState().TeamTerrorists().Members() {
		if ply.IsAlive() && ply.SteamID64 != victimSteamId {
			aliveT++
		}
	}
	for _, ply := range p.GameState().TeamCounterTerrorists().Members() {
		if ply.IsAlive() && ply.SteamID64 != victimSteamId {
			aliveCT++
		}
	}
	return aliveT, aliveCT
}

//Простая загрузка URL методом GET
func getUrl(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		panic("Возможно проблема с сетью? Не могу загрузить" + url)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return bodyBytes
}

//Скачивание данных по URL с указанными параметрами в заголовке
func urlRequest(url string, method string, urlParams map[string]string) []byte {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic("Возможно проблема с сетью? Не могу загрузить" + url)
	}

	for k, v := range urlParams {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic("Возможно проблема с сетью? Не могу загрузить" + url)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	return bodyBytes
}

//Парсинг byte[] полученного по сети в разбираемый JSON-объект
func parseJson(jsonBytes []byte) map[string]interface{} {
	var ret map[string]interface{}
	err = json.Unmarshal(jsonBytes, &ret)
	if err != nil {
		panic(err)
	}
	return ret
}

//Загрузка файла по сети
func downloadFile(filepath string, url string, method string, urlParams map[string]string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}

	for k, v := range urlParams {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}

func settingsLoad() {
	f, err := os.Open("settings.json")
	if err != nil {
		settingsSave()
		return
	}
	settingsBytes, _ := ioutil.ReadAll(f)
	json.Unmarshal(settingsBytes, &settings)
	defer f.Close()
}

func settingsSave() {
	settingsBytes, _ := json.MarshalIndent(settings, "", "\t")
	settingsStr := string(settingsBytes)
	print(settingsStr)
	os.Remove("settings.json")
	f, err := os.Create("settings.json")
	if err != nil {
		panic("Ошибка при загрузке файла настроек")
	}
	f.Write(settingsBytes)
	defer f.Close()
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
