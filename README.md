# gitStats


Консольная утилита для подсчёта статистики авторов git репозитория.

✗ gitfame --repository=. --extensions='.go,.md' --order-by=lines \newline
Name                   Lines Commits Files
Joe Tsai               12154 92      49
colinnewell            130   1       1
Roger Peppe            59    1       2
A. Ishikawa            36    1       1
Tobias Klauser         33    1       2
178inaba               11    2       4
Kyle Lemons            11    1       1
Dmitri Shuralyov       8     1       2
ferhat elmas           7     1       4
Christian Muehlhaeuser 6     3       4
k.nakada               5     1       3
LMMilewski             5     1       2
Ernest Galbrun         3     1       1
Ross Light             2     1       1
Chris Morrow           1     1       1
Fiisio                 1     1       1

Флаги

--repository — путь до Git репозитория; по умолчанию текущая директория

--revision — указатель на коммит; HEAD по умолчанию

--order-by — ключ сортировки результатов; один из lines (дефолт), commits, files.

По умолчанию результаты сортируются по убыванию ключа (lines, commits, files). При равенстве ключей выше будет автор с лексикографически меньшим именем. При использовании флага соответствующее поле в ключе перемещается на первое место.

--use-committer — булев флаг, заменяющий в расчётах автора (дефолт) на коммиттера

--format — формат вывода; один из tabular (дефолт), csv, json, json-lines;

tabular:

Name         Lines Commits Files
Joe Tsai     64    3       2
Ross Light   2     1       1
ferhat elmas 1     1       1
Human-readable формат. Для паддинга используется пробел. см. text/tabwriter.

csv:

Name,Lines,Commits,Files
Joe Tsai,64,3,2
Ross Light,2,1,1
ferhat elmas,1,1,1
encoding/csv

json:

[{"name":"Joe Tsai","lines":64,"commits":3,"files":2},{"name":"Ross Light","lines":2,"commits":1,"files":1},{"name":"ferhat elmas","lines":1,"commits":1,"files":1}]
encoding/json

json-lines:

{"name":"Joe Tsai","lines":64,"commits":3,"files":2}
{"name":"Ross Light","lines":2,"commits":1,"files":1}
{"name":"ferhat elmas","lines":1,"commits":1,"files":1}
--extensions — список расширений, сужающий список файлов в расчёте; множество ограничений разделяется запятыми, например, '.go,.md'

--languages — список языков (программирования, разметки и др.), сужающий список файлов в расчёте; множество ограничений разделяется запятыми, например 'go,markdown'

--exclude — набор Glob паттернов, исключающих файлы из расчёта, например 'foo/*,bar/*'

--restrict-to — набор Glob паттернов, исключающий все файлы, не удовлетворяющие ни одному из паттернов набора
