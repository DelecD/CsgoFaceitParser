Данное приложение в автоматическом режиме скачивает и анализирует матчи из турниров Faceit по CS:GO. В результате анализа в БД загружаются искомые события (kills, bomb planting, bomb defusing).
Используется парсер: github.com/markus-wa/demoinfocs-golang

Парсер хранит загруженное в папке "demos". Перед стартом нужно поправить settings.json (особенно в плане настроек MySQL и залить в базу таблицы (все нужные запросы для создания таблиц в demos.sql).

Настройки парсера в settings.json:
- CompetitionIds, PagesRange, PageSize - это аргументы, которые парсер подставляет в ссылки для скачивания вида https://api.faceit.com/match-history/v4/matches/competition?id=42e160fc-2651-4fa5-9a9b-829199e27adb&page=1000&size=0&type=matchmaking. 
CompetitionIds - айдишники соревнований
PagesRange - первое значение - начальная страница, второе - конечная.
PagesSize - сколько демок на страницу.

Если
CompetitionIds = ["abc", "DEF"]
PagesRange = [1,3]
PageSize = 50,

то будут обработаны страницы
https://api.faceit.com/match-history/v4/matches/competitionid=abc&page=1&size=50&type=matchmaking
https://api.faceit.com/match-history/v4/matches/competitionid=abc&page=2&size=50&type=matchmaking
https://api.faceit.com/match-history/v4/matches/competitionid=abc&page=3&size=50&type=matchmaking
https://api.faceit.com/match-history/v4/matches/competitionid=DEF&page=1&size=50&type=matchmaking
https://api.faceit.com/match-history/v4/matches/competitionid=DEF&page=2&size=50&type=matchmaking
https://api.faceit.com/match-history/v4/matches/competitionid=DEF&page=3&size=50&type=matchmaking

- Verbose - нужно ли парсеру выводить статус на экран
- MatchesParallels - сколько параллельно обрабатывается матчей (загрузка архива, распаковка, парсинг). При значениях около 20 FaceIT начинает ругаться, на 10 работало нормально
- WaitIf403ErrorSecs - сколько секунд ждать перед повторной попыткой скачать демку, если ошибка 403 (такие ошибки случаются регулярно в начале работы, в консоль при этом пишет сообщение "bad request", через небольшое время они проходят)
- SaveDemoFiles - если false, то архив демки будет удаляться после того, как демка обработана
- FaceitApiKey и параметры mysql - думаю, тут понятно :)

main.go - тут исходники
build.bat - компилирующий обе версии батник
main.exe, main - сборки под windows и linux соответственно, результат работы build.bat
demos.sql - файл инициализации таблиц БД
go.mod, go.sum - служебные (там указывается, какие библиотеки надо подтягивать и прочее)