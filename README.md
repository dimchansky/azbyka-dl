# azbyka-dl [![GoDoc][1]][2] [![Build Status][3]][4] [![Go Report Card][5]][6] [![Coverage Status][7]][8]
            
[1]: https://godoc.org/github.com/dimchansky/azbyka-dl?status.svg
[2]: https://godoc.org/github.com/dimchansky/azbyka-dl
[3]: https://travis-ci.org/dimchansky/azbyka-dl.svg?branch=master
[4]: https://travis-ci.org/dimchansky/azbyka-dl
[5]: https://goreportcard.com/badge/github.com/dimchansky/azbyka-dl
[6]: https://goreportcard.com/report/github.com/dimchansky/azbyka-dl
[7]: https://codecov.io/gh/dimchansky/azbyka-dl/branch/master/graph/badge.svg
[8]: https://codecov.io/gh/dimchansky/azbyka-dl

Сайт https://azbyka.ru/audio/ позволяет скачивать либо плейлист в формате M3U, либо сразу всю книгу в формате M4B. 
Однако мне хотелось скачивать книгу в виде набора из mp3-файлов. К сожалению, делать это через браузер крайне неудобно,
когда книга состоит из сотни отдельных mp3-файлов. Чтобы облегчить этот процесс, написал утилиту, которая 
автоматически собирает список mp3-файлов со страницы книги, скачивает их поочередно, добавляя нумерацию и изменяя 
при этом название файла на тот, как трек назывался на самом сайте.

Так как изначально писал утилиту для себя, то сделал это максимально просто, в виде программы без графического интерфейса,
которая запускается из командной строки. Т.к. позже заметил в комментариях на том же сайте, что люди спрашивают об этой 
функциональности, то решил с минимальными изменениями поделиться утилитой с другими, возможно кому-то пригодится.

## Инструкция по использованию

Первое, что нужно сделать - это скачать саму утилиту со страницы [releases](/releases). Утилита собрана для трех 
операционных систем (Windows, MacOS, Linux), поэтому скачивать нужно архив, в названии которого присутствует нужная
операционная система. Пользоваться относительно несложно, нужно просто указать адрес страницы с аудио-книгой, откуда 
нужно скачать mp3-файлы.

Например, команда

    azbyka-dl https://azbyka.ru/audio/zhizneopisanie-i-duhovnoe-nasledie-protoiereja-pontija-rupysheva.html

скачает все главы со страницы [Жизнеописание и духовное наследие протоиерея Понтия Рупышева](https://azbyka.ru/audio/zhizneopisanie-i-duhovnoe-nasledie-protoiereja-pontija-rupysheva.html)
в виде отдельных mp3-файлов в текущий каталог.

Если хочется скачать файлы не в текущий каталог, а в другой, то это можно сделать с помощью опции `-dir`:

    azbyka-dl -dir ~/mp3/pontij-rupyshev https://azbyka.ru/audio/zhizneopisanie-i-duhovnoe-nasledie-protoiereja-pontija-rupysheva.html
    
Команда выше скачает все файлы в каталог `~/mp3/pontij-rupyshev`.

Есть еще опция `-skip`, которая позволяет пропустить скачивание нескольких файлов, позволяет продолжить скачивание, например,
после потери соединения с Интернетом. Например, команда

    azbyka-dl -skip 3 https://azbyka.ru/audio/zhizneopisanie-i-duhovnoe-nasledie-protoiereja-pontija-rupysheva.html
    
пропустит первые 3 файла и продолжит скачивание mp3-файлов с 4-ого файла в текущий каталог.

Если возникнут какие-то проблемы при использовании, пишите о них на странице [issues](/issues)
(кнопка [`New issue`](/issues/new)).