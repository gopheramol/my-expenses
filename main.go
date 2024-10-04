package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Expense struct {
	ID       int     `json:"id"`
	Amount   float64 `json:"amount"`
	Date     string  `json:"date"`
	Category string  `json:"category"`
	Note     string  `json:"note"`
}

var expenses []Expense

func main() {
	loadExpenses()

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/expenses", handleExpenses)
	http.HandleFunc("/api/expense", handleExpense)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func loadExpenses() {
	file, err := ioutil.ReadFile("expenses.json")
	if err != nil {
		expenses = []Expense{}
		return
	}

	err = json.Unmarshal(file, &expenses)
	if err != nil {
		log.Fatal(err)
	}
}

func saveExpenses() {
	file, err := json.MarshalIndent(expenses, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("expenses.json", file, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleExpenses(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		json.NewEncoder(w).Encode(expenses)
	case "POST":
		var expense Expense
		err := json.NewDecoder(r.Body).Decode(&expense)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		expense.ID = len(expenses) + 1
		expenses = append(expenses, expense)
		saveExpenses()

		json.NewEncoder(w).Encode(expense)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleExpense(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "PUT":
		var updatedExpense Expense
		err := json.NewDecoder(r.Body).Decode(&updatedExpense)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for i, expense := range expenses {
			if expense.ID == id {
				updatedExpense.ID = id
				expenses[i] = updatedExpense
				saveExpenses()
				json.NewEncoder(w).Encode(updatedExpense)
				return
			}
		}

		http.Error(w, "Expense not found", http.StatusNotFound)
	case "DELETE":
		for i, expense := range expenses {
			if expense.ID == id {
				expenses = append(expenses[:i], expenses[i+1:]...)
				saveExpenses()
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		http.Error(w, "Expense not found", http.StatusNotFound)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
