# expvar for gin

## EOL-warning

**This package has been abandoned on 2016-12-15. Please use [gin-contrib/expvar](https://github.com/gin-contrib/expvar) instead.**

```go
package main

import "github.com/gin-gonic/gin"
import "github.com/gin-gonic/contrib/expvar"

func main() {
	router := gin.Default()
	router.GET("/debug/vars", expvar.Handler())
	router.Run(":8080")
}
```

Request: `http://localhost:8080/debug/vars`

```json
{
"cmdline": ["/var/folders/zg/q__7tncn7kxc34pc3j0_v7rc0000gn/T/go-build115008999/command-line-arguments/_obj/exe/main"],
"memstats": {"Alloc":406968,"TotalAlloc":3025088,"Sys":3999992,"Lookups":15,"Mallocs":10405,"Frees":9228,"HeapAlloc":406968,"HeapSys":1916928,"HeapIdle":1097728,"HeapInuse":819200,"HeapReleased":0,"HeapObjects":1177,"StackInuse":180224,"StackSys":180224,"MSpanInuse":7904,"MSpanSys":16384,"MCacheInuse":1200,"MCacheSys":16384,"BuckHashSys":1441424,"GCSys":137622,"OtherSys":291026,"NextGC":578160,"LastGC":1432428478423798618,"PauseTotalNs":4326675,"PauseNs":[105780,76617,91326,115727,195752,249831,554025,485129,344607,416552,400537,424639,463089,403064,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"PauseEnd":[1432428359211210066,1432428359211415512,1432428359211821428,1432428359212452447,1432428359212906768,1432428359214282772,1432428381213398837,1432428383977712300,1432428452748952359,1432428470574839824,1432428472452814302,1432428474379491025,1432428476329036668,1432428478423798444,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"NumGC":14,"EnableGC":true,"DebugGC":false,"BySize":[{"Size":0,"Mallocs":0,"Frees":0},{"Size":8,"Mallocs":258,"Frees":237},{"Size":16,"Mallocs":4125,"Frees":3670},{"Size":32,"Mallocs":765,"Frees":652},{"Size":48,"Mallocs":495,"Frees":367},{"Size":64,"Mallocs":305,"Frees":274},{"Size":80,"Mallocs":37,"Frees":24},{"Size":96,"Mallocs":32,"Frees":25},{"Size":112,"Mallocs":713,"Frees":569},{"Size":128,"Mallocs":283,"Frees":259},{"Size":144,"Mallocs":91,"Frees":85},{"Size":160,"Mallocs":139,"Frees":112},{"Size":176,"Mallocs":349,"Frees":315},{"Size":192,"Mallocs":0,"Frees":0},{"Size":208,"Mallocs":223,"Frees":196},{"Size":224,"Mallocs":2,"Frees":2},{"Size":240,"Mallocs":86,"Frees":80},{"Size":256,"Mallocs":21,"Frees":18},{"Size":288,"Mallocs":130,"Frees":104},{"Size":320,"Mallocs":19,"Frees":15},{"Size":352,"Mallocs":278,"Frees":259},{"Size":384,"Mallocs":5,"Frees":5},{"Size":416,"Mallocs":17,"Frees":10},{"Size":448,"Mallocs":0,"Frees":0},{"Size":480,"Mallocs":2,"Frees":1},{"Size":512,"Mallocs":16,"Frees":12},{"Size":576,"Mallocs":92,"Frees":83},{"Size":640,"Mallocs":12,"Frees":7},{"Size":704,"Mallocs":2,"Frees":2},{"Size":768,"Mallocs":0,"Frees":0},{"Size":896,"Mallocs":15,"Frees":12},{"Size":1024,"Mallocs":10,"Frees":8},{"Size":1152,"Mallocs":88,"Frees":82},{"Size":1280,"Mallocs":7,"Frees":6},{"Size":1408,"Mallocs":2,"Frees":1},{"Size":1536,"Mallocs":0,"Frees":0},{"Size":1664,"Mallocs":9,"Frees":4},{"Size":2048,"Mallocs":10,"Frees":9},{"Size":2304,"Mallocs":88,"Frees":82},{"Size":2560,"Mallocs":6,"Frees":5},{"Size":2816,"Mallocs":2,"Frees":1},{"Size":3072,"Mallocs":1,"Frees":1},{"Size":3328,"Mallocs":4,"Frees":1},{"Size":4096,"Mallocs":172,"Frees":167},{"Size":4608,"Mallocs":93,"Frees":81},{"Size":5376,"Mallocs":7,"Frees":0},{"Size":6144,"Mallocs":87,"Frees":81},{"Size":6400,"Mallocs":0,"Frees":0},{"Size":6656,"Mallocs":1,"Frees":0},{"Size":6912,"Mallocs":0,"Frees":0},{"Size":8192,"Mallocs":6,"Frees":6},{"Size":8448,"Mallocs":0,"Frees":0},{"Size":8704,"Mallocs":1,"Frees":1},{"Size":9472,"Mallocs":0,"Frees":0},{"Size":10496,"Mallocs":0,"Frees":0},{"Size":12288,"Mallocs":1,"Frees":1},{"Size":13568,"Mallocs":0,"Frees":0},{"Size":14080,"Mallocs":0,"Frees":0},{"Size":16384,"Mallocs":0,"Frees":0},{"Size":16640,"Mallocs":0,"Frees":0},{"Size":17664,"Mallocs":1,"Frees":0}]}
}
```
