build:
	go build -o GemBoard github.com/JamesMoreau/GemBoard

buildWeb:
	gogio -target js github.com/JamesMoreau/GemBoard 

runWeb:
	goexec 'http.ListenAndServe(":8080", http.FileServer(http.Dir("GemBoard")))'