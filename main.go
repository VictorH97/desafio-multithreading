package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

type ResponseCep struct {
	API  string
	Data string
}

func main() {
	http.HandleFunc("/", BuscaCepHandler)
	http.ListenAndServe(":8080", nil)
}

func BuscaCepHandler(w http.ResponseWriter, r *http.Request) {
	responseCep := make(chan ResponseCep)
	cepParam := r.URL.Query().Get("cep")

	valid, err := regexp.MatchString("\\d{5}-*\\d{3}", cepParam)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, "CEP inválido", http.StatusInternalServerError)
		return
	}

	go buscarCepBrasilAPI(cepParam, responseCep)
	go buscarCepViaCep(cepParam, responseCep)

	select {
	case response := <-responseCep:
		fmt.Printf("API: %s, Dados: %s\n", response.API, response.Data)
		w.WriteHeader(http.StatusOK)
	case <-time.After(time.Second):
		http.Error(w, "Timeout", http.StatusRequestTimeout)
	}
}

func buscarCepBrasilAPI(cep string, cepData chan<- ResponseCep) {
	resp, err := http.Get("https://brasilapi.com.br/api/cep/v1/" + cep)

	if err != nil {
		println(err.Error())
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		println(err.Error())
		return
	}

	cepData <- ResponseCep{"BrasilAPI", string(body)}

	return
}

func buscarCepViaCep(cep string, cepData chan<- ResponseCep) {
	resp, err := http.Get("http://viacep.com.br/ws/" + cep + "/json/")

	if err != nil {
		println(err.Error())
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		println(err.Error())
		return
	}

	cepData <- ResponseCep{"ViaCep", string(body)}

	return
}
