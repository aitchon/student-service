package repositories

import (
	"database/sql"
	"errors"
	"student-service/models"
)

var ErrStudentNotFound = errors.New("student not found")
var ErrDuplicateName = errors.New("student with this name already exists")

type StudentRepository struct {
	db *sql.DB
}

func NewStudentRepository(db *sql.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

func (r *StudentRepository) GetAllStudents(name string) ([]models.Student, error) {
	query := "SELECT id, name FROM students"
	args := []interface{}{}

	if name != "" {
		query += " WHERE name = ?"
		args = append(args, name)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var student models.Student
		if err := rows.Scan(&student.ID, &student.Name); err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	return students, nil
}

func (r *StudentRepository) GetStudentByID(id int) (*models.Student, error) {
	student := &models.Student{}
	err := r.db.QueryRow("SELECT id, name FROM students WHERE id = ?", id).
		Scan(&student.ID, &student.Name)
	if err == sql.ErrNoRows {
		return nil, ErrStudentNotFound
	}
	if err != nil {
		return nil, err
	}
	return student, nil
}

func (r *StudentRepository) CreateStudent(student *models.Student) error {
	result, err := r.db.Exec("INSERT INTO students (name) VALUES (?)",
		student.Name)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	student.ID = int(id)
	return nil
}

func (r *StudentRepository) UpdateStudent(student *models.Student) error {
	result, err := r.db.Exec("UPDATE students SET name = ? WHERE id = ?",
		student.Name, student.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrStudentNotFound
	}
	return nil
}

func (r *StudentRepository) DeleteStudent(id int) error {
	result, err := r.db.Exec("DELETE FROM students WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrStudentNotFound
	}
	return nil
}
