EANN direktorijā ir viss projekts.
Tajā ir divi direktoriji: AI un Game.
AI ir paketes, kas apraksta neironu tīklus un evolūcijas algoritmus.
Game direktorijā ir visi galvenie faili un viena eksperimenta rezultāti.
Stats direktorijā ir eksperimenta laikā savāktā statistika, un Trained direktorijā ir faili ar apmācītiem neironu tīkliem.
main.go satur apmācības kodu. game.go satur programmu, kas simulē spēli ar katru aģentu. Lai šī programma darbotos, ir nepieciešama direktorija Trained ar apmācītiem neironu tīkliem. GraphGen.go satur kodus, lai ģenerētu grafikus, pamatojoties uz savākto statistiku, vai var iegūt maksimālās vērtības katram aģentam un apkopotās vidējās vērtības. Lai šī programma darbotos, ir nepieciešama direktorija Stats.

Lai palaistu game.exe, tajā pašā direktorijā ir nepieciešama SDL2.dll. https://github.com/libsdl-org/SDL/releases.

Lai kompilētu main.go un GraphGen.go, ir nepieciešama Golang programmēšanas valoda. https://go.dev/dl/
Lai kompilētu game.go, jums būs nepieciešama arī sdl instalācija. https://github.com/veandco/go-sdl2
