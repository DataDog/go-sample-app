package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	_ "modernc.org/sqlite"
)

func PostFormWithContext(ctx context.Context, c *http.Client, url string, data url.Values) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.Do(req)
}

func GetWithContext(ctx context.Context, c *http.Client, url string) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func main() {
	db, err := sql.Open("sqlite", "./data/userdb.sqlite")
	if err != nil {
		log.Fatalf("Failed to open the user database: %v", err)
	}

	// set up database
	db.ExecContext(context.Background(),
		`CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email STRING UNIQUE,
	fname STRING,
	lname STRING
)`)

	db.ExecContext(context.Background(),
		`INSERT OR REPLACE INTO users(email, fname, lname)
VALUES ('fake.email@somecompany.com', 'John', 'Smith'),
		('alice@othercompany.com', 'Alice', 'Rivest'),
		('bob@thirdcompany.com', 'Bob', 'Shamir');
`)

	var c http.Client

	http.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/user/")
		id = strings.SplitN(id, "/", 2)[0]

		rows, err := db.QueryContext(r.Context(), "SELECT email, fname, lname FROM users WHERE id = ?;", id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to query db: %v", err), http.StatusInternalServerError)
			return
		}
		var user struct {
			Email string
			Fname string
			Lname string
		}

		if r.Method == http.MethodPost {
			var found bool
			for rows.Next() {
				if found {
					http.Error(w, fmt.Sprintf("Query returned more than one user for id %s.", id), http.StatusInternalServerError)
					return
				}
				found = true
			}
			if !found {
				http.Error(w, "No such user.", http.StatusBadRequest)
				return
			}

			r.ParseForm()
			content := strings.TrimSpace(r.FormValue("note"))
			if content == "" {
				http.Error(w, "Cannot submit an empty note.", http.StatusBadRequest)
				return
			}
			resp, err := PostFormWithContext(r.Context(), &c, "http://localhost:8081/new", url.Values{"userid": {id}, "content": {content}})
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to submit note: %v", err), http.StatusInternalServerError)
				return
			}
			if resp.StatusCode != 200 {
				body, err := io.ReadAll(resp.Body)
				defer resp.Body.Close()
				if err != nil {
					http.Error(w, fmt.Sprintf("Failed to read error response from note submission: %v", err), http.StatusInternalServerError)
					return
				}
				http.Error(w, fmt.Sprintf("Failed to submit note: %v", string(body)), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, r.URL.String(), http.StatusFound)
		}

		var found bool
		for rows.Next() {
			if found {
				http.Error(w, fmt.Sprintf("Query returned more than one user for id %s.", id), http.StatusInternalServerError)
				return
			}
			found = true
			rows.Scan(&user.Email, &user.Fname, &user.Lname)
			fmt.Printf("USER: %#v\n", user)
			fmt.Fprintf(w, "<p>User: %v</p>", user)
			fmt.Fprintf(w, `<form action="" method="post"><textarea name="note" rows="24" cols="80"></textarea><p><input type="submit" value="Submit Note"/></p></form>`)

			resp, err := GetWithContext(r.Context(), &c, fmt.Sprintf("http://localhost:8081/notes?userid=%s", id))
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to query notes for user %s: %v", id, err), http.StatusInternalServerError)
				return
			}
			dec := json.NewDecoder(resp.Body)
			defer resp.Body.Close()
			for {
				var note struct {
					ID      int
					Content string
					Created string
				}
				err := dec.Decode(&note)
				if err != nil {
					if err != io.EOF {
						http.Error(w, fmt.Sprintf("Failed to decode notes for user %s: %v", id, err), http.StatusInternalServerError)
					}
					break
				}
				fmt.Fprintf(w, "<p>Notes</p><table>")
				fmt.Fprintf(w, "<tr><th>ID</th><th>Creation Time</th><th>Note</th></tr>")
				fmt.Fprintf(w, "<tr><td>%d</td><td>%s</td><td>%s</td></tr>", note.ID, note.Created, note.Content)
				fmt.Fprintf(w, "</table>")
			}
		}
		if err := rows.Err(); err != nil {
			http.Error(w, fmt.Sprintf("Failed to query db: %v", err), http.StatusInternalServerError)
			return
		}
		if !found {
			fmt.Fprintf(w, "No such user.")
		}

	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/user/")
		id = strings.SplitN(id, "/", 2)[0]
		rows, err := db.QueryContext(r.Context(), "SELECT id, email FROM users;")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to query db: %v", err), http.StatusInternalServerError)
			return
		}
		var user struct {
			ID    int
			Email string
		}
		fmt.Fprintf(w, "<table>")
		for rows.Next() {
			rows.Scan(&user.ID, &user.Email)
			fmt.Printf("USER: %#v\n", user)
			fmt.Fprintf(w, `<tr><td><a href="/user/%d">%v</a></td><td>%v</td></tr>`, user.ID, user.ID, user.Email)
		}
		fmt.Fprintf(w, "</table>")
		if err := rows.Err(); err != nil {
			http.Error(w, fmt.Sprintf("Failed to query db: %v", err), http.StatusInternalServerError)
			return
		}
	})

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
