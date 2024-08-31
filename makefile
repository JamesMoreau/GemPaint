
buildWeb:
	gogio -target js github.com/JamesMoreau/GemPaint 

runWeb:
	goexec 'http.ListenAndServe(":8080", http.FileServer(http.Dir("GemPaint")))'