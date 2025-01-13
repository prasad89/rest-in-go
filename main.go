package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Task struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "tasks.db")
	if err != nil {
		return err
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		status TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func getTasks(c *gin.Context) {
	rows, err := db.Query("SELECT id, title, status FROM tasks")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch tasks",
		})
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to scan task",
			})
			return
		}
		tasks = append(tasks, task)
	}
	c.JSON(http.StatusOK, tasks)
}

func getTask(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid task ID",
		})
		return
	}

	var task Task
	err = db.QueryRow("SELECT id, title, status FROM tasks WHERE id = ?", taskID).Scan(
		&task.ID, &task.Title, &task.Status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "task not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to fetch task",
			})
		}
		return
	}

	c.JSON(http.StatusOK, task)
}

func createTask(c *gin.Context) {
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON input",
		})
		return
	}

	result, err := db.Exec("INSERT INTO tasks (title, status) VALUES (?, ?)", task.Title, task.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create task",
		})
		return
	}

	taskID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get task ID",
		})
		return
	}
	task.ID = int(taskID)

	c.JSON(http.StatusCreated, task)
}

func updateTask(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid task ID",
		})
		return
	}

	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON input",
		})
		return
	}

	result, err := db.Exec("UPDATE tasks SET title = ?, status = ? WHERE id = ?", task.Title, task.Status, taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update task",
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "task not found or no changes made",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "task updated successfully",
	})
}

func deleteTask(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid task ID",
		})
		return
	}

	result, err := db.Exec("DELETE FROM tasks WHERE id = ?", taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete task",
		})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "task not found or no changes made",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "task deleted succesfully",
	})
}

func main() {
	err := initDB()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	router := gin.Default()

	router.GET("/ping", ping)
	router.GET("/tasks", getTasks)
	router.GET("/task/:id", getTask)
	router.POST("/task", createTask)
	router.PUT("/task/:id", updateTask)
	router.DELETE("/task/:id", deleteTask)

	router.Run()
}
