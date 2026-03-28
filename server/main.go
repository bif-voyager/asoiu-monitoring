package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/cors"
	_ "modernc.org/sqlite"
)

var db *sql.DB

func initDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		return err
	}

	tables := []string{
		`CREATE TABLE IF NOT EXISTS students (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			fio TEXT NOT NULL,
			record_book_number TEXT NOT NULL,
			group_name TEXT NOT NULL,
			course INTEGER NOT NULL DEFAULT 1,
			study_form TEXT NOT NULL DEFAULT 'Очная',
			training_direction TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'Активен'
		);`,
		`CREATE TABLE IF NOT EXISTS teachers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			fio TEXT NOT NULL,
			position TEXT NOT NULL DEFAULT '',
			academic_degree TEXT NOT NULL DEFAULT '',
			login TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS disciplines (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			semester INTEGER NOT NULL DEFAULT 1,
			control_type TEXT NOT NULL DEFAULT 'Экзамен',
			hours INTEGER NOT NULL DEFAULT 0,
			teacher_id INTEGER,
			FOREIGN KEY(teacher_id) REFERENCES teachers(id)
		);`,
		`CREATE TABLE IF NOT EXISTS performance (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			discipline_id INTEGER NOT NULL,
			teacher_id INTEGER,
			sheet_number TEXT NOT NULL DEFAULT '',
			sheet_fill_date TEXT NOT NULL DEFAULT '',
			grade TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT '',
			comment TEXT NOT NULL DEFAULT '',
			has_debt INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT '',
			FOREIGN KEY(student_id) REFERENCES students(id),
			FOREIGN KEY(discipline_id) REFERENCES disciplines(id),
			FOREIGN KEY(teacher_id) REFERENCES teachers(id)
		);`,
		`CREATE TABLE IF NOT EXISTS commissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			student_id INTEGER NOT NULL,
			performance_id INTEGER NOT NULL,
			assigned_date TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'Назначена',
			decision TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			FOREIGN KEY(student_id) REFERENCES students(id),
			FOREIGN KEY(performance_id) REFERENCES performance(id)
		);`,
	}

	for _, query := range tables {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("create table error: %w", err)
		}
	}

	// Seed demo data if tables are empty
	var count int
	db.QueryRow("SELECT COUNT(*) FROM teachers").Scan(&count)
	if count == 0 {
		seedData()
	}

	return nil
}

func hashPassword(password string) string {
	h := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", h)
}

func seedData() {
	now := time.Now().Format("2006-01-02")

	// Teachers
	teachers := []struct {
		fio, position, degree, login, password string
	}{
		{"Иванов Сергей Петрович", "Профессор", "Доктор наук", "ivanov", "123456"},
		{"Петрова Елена Алексеевна", "Доцент", "Кандидат наук", "petrova", "123456"},
		{"Сидоров Дмитрий Николаевич", "Старший преподаватель", "Кандидат наук", "sidorov", "123456"},
		{"Козлова Анна Михайловна", "Доцент", "Кандидат наук", "kozlova", "123456"},
		{"Новиков Алексей Владимирович", "Профессор", "Доктор наук", "novikov", "123456"},
	}
	for _, t := range teachers {
		db.Exec("INSERT INTO teachers (fio, position, academic_degree, login, password_hash) VALUES (?,?,?,?,?)",
			t.fio, t.position, t.degree, t.login, hashPassword(t.password))
	}

	// Students — 3 groups × 5 students
	students := []struct {
		fio, recordBook, group string
		course                 int
		studyForm, direction   string
	}{
		{"Абрамов Иван Сергеевич", "2024-001", "ИС-21", 2, "Очная", "Информационные системы"},
		{"Белова Мария Дмитриевна", "2024-002", "ИС-21", 2, "Очная", "Информационные системы"},
		{"Васильев Пётр Андреевич", "2024-003", "ИС-21", 2, "Очная", "Информационные системы"},
		{"Григорьева Анна Олеговна", "2024-004", "ИС-21", 2, "Очная", "Информационные системы"},
		{"Дмитриев Олег Игоревич", "2024-005", "ИС-21", 2, "Очная", "Информационные системы"},

		{"Егоров Максим Викторович", "2024-006", "ПИ-31", 3, "Очная", "Программная инженерия"},
		{"Жукова Елена Сергеевна", "2024-007", "ПИ-31", 3, "Очная", "Программная инженерия"},
		{"Зайцев Артём Павлович", "2024-008", "ПИ-31", 3, "Очная", "Программная инженерия"},
		{"Ильина Дарья Романовна", "2024-009", "ПИ-31", 3, "Очная", "Программная инженерия"},
		{"Кузнецов Никита Алексеевич", "2024-010", "ПИ-31", 3, "Очная", "Программная инженерия"},

		{"Лебедев Роман Дмитриевич", "2024-011", "КБ-41", 4, "Очная", "Кибербезопасность"},
		{"Морозова Ольга Ивановна", "2024-012", "КБ-41", 4, "Очная", "Кибербезопасность"},
		{"Николаев Сергей Владимирович", "2024-013", "КБ-41", 4, "Очная", "Кибербезопасность"},
		{"Орлова Виктория Андреевна", "2024-014", "КБ-41", 4, "Очная", "Кибербезопасность"},
		{"Павлов Денис Олегович", "2024-015", "КБ-41", 4, "Очно-заочная", "Кибербезопасность"},
	}
	for _, s := range students {
		db.Exec("INSERT INTO students (fio, record_book_number, group_name, course, study_form, training_direction, status) VALUES (?,?,?,?,?,?,?)",
			s.fio, s.recordBook, s.group, s.course, s.studyForm, s.direction, "Активен")
	}

	// Disciplines
	disciplines := []struct {
		name, controlType string
		semester, hours   int
		teacherId         int
	}{
		{"Базы данных", "Экзамен", 3, 108, 1},
		{"Программирование на Go", "Экзамен", 4, 72, 1},
		{"Математический анализ", "Экзамен", 3, 144, 2},
		{"Операционные системы", "Зачёт", 5, 90, 3},
		{"Компьютерные сети", "Экзамен", 6, 108, 4},
		{"Информационная безопасность", "Зачёт", 7, 72, 5},
	}
	for _, d := range disciplines {
		db.Exec("INSERT INTO disciplines (name, semester, control_type, hours, teacher_id) VALUES (?,?,?,?,?)",
			d.name, d.semester, d.controlType, d.hours, d.teacherId)
	}

	// Performance records
	type perfRec struct {
		studentId, disciplineId, teacherId int
		sheetNum, grade, status, comment   string
		hasDebt                            int
	}
	perfRecords := []perfRec{
		// ИС-21 group for "Базы данных" (discipline 1, teacher 1)
		{1, 1, 1, "ВД-001", "Отлично", "Сдано", "", 0},
		{2, 1, 1, "ВД-001", "Хорошо", "Сдано", "", 0},
		{3, 1, 1, "ВД-001", "Неудовлетворительно", "Не сдано", "Не явился на экзамен", 1},
		{4, 1, 1, "ВД-001", "Удовлетворительно", "Сдано", "", 0},
		{5, 1, 1, "ВД-001", "Неудовлетворительно", "Не сдано", "Плохая подготовка", 1},

		// ИС-21 group for "Математический анализ" (discipline 3, teacher 2)
		{1, 3, 2, "ВД-002", "Хорошо", "Сдано", "", 0},
		{2, 3, 2, "ВД-002", "Отлично", "Сдано", "", 0},
		{3, 3, 2, "ВД-002", "Неудовлетворительно", "Не сдано", "Систематические пропуски", 1},
		{4, 3, 2, "ВД-002", "Хорошо", "Сдано", "", 0},
		{5, 3, 2, "ВД-002", "Удовлетворительно", "Сдано", "", 0},

		// ПИ-31 group for "Операционные системы" (discipline 4, teacher 3)
		{6, 4, 3, "ВД-003", "Зачтено", "Сдано", "", 0},
		{7, 4, 3, "ВД-003", "Зачтено", "Сдано", "", 0},
		{8, 4, 3, "ВД-003", "Не зачтено", "Не сдано", "Не сдал лабораторные работы", 1},
		{9, 4, 3, "ВД-003", "Зачтено", "Сдано", "", 0},
		{10, 4, 3, "ВД-003", "Не зачтено", "Не сдано", "Низкая посещаемость", 1},

		// ПИ-31 group for "Программирование на Go" (discipline 2, teacher 1)
		{6, 2, 1, "ВД-004", "Отлично", "Сдано", "", 0},
		{7, 2, 1, "ВД-004", "Хорошо", "Сдано", "", 0},
		{8, 2, 1, "ВД-004", "Удовлетворительно", "Сдано", "", 0},
		{9, 2, 1, "ВД-004", "Неудовлетворительно", "Не сдано", "Не сдал проект", 1},
		{10, 2, 1, "ВД-004", "Хорошо", "Сдано", "", 0},

		// КБ-41 group for "Компьютерные сети" (discipline 5, teacher 4)
		{11, 5, 4, "ВД-005", "Отлично", "Сдано", "", 0},
		{12, 5, 4, "ВД-005", "Хорошо", "Сдано", "", 0},
		{13, 5, 4, "ВД-005", "Удовлетворительно", "Сдано", "", 0},
		{14, 5, 4, "ВД-005", "Неудовлетворительно", "Не сдано", "Не выполнил курсовую работу", 1},
		{15, 5, 4, "ВД-005", "Удовлетворительно", "Сдано", "", 0},

		// КБ-41 group for "Информационная безопасность" (discipline 6, teacher 5)
		{11, 6, 5, "ВД-006", "Зачтено", "Сдано", "", 0},
		{12, 6, 5, "ВД-006", "Не зачтено", "Не сдано", "Не защитил лабораторные", 1},
		{13, 6, 5, "ВД-006", "Зачтено", "Сдано", "", 0},
		{14, 6, 5, "ВД-006", "Не зачтено", "Не сдано", "Пропуски занятий", 1},
		{15, 6, 5, "ВД-006", "Зачтено", "Сдано", "", 0},
	}
	for _, p := range perfRecords {
		db.Exec(`INSERT INTO performance (student_id, discipline_id, teacher_id, sheet_number, sheet_fill_date, grade, status, comment, has_debt, created_at)
			VALUES (?,?,?,?,?,?,?,?,?,?)`,
			p.studentId, p.disciplineId, p.teacherId, p.sheetNum, now, p.grade, p.status, p.comment, p.hasDebt, now)
	}

	// Commissions
	db.Exec(`INSERT INTO commissions (student_id, performance_id, assigned_date, status, decision, description)
		VALUES (?,?,?,?,?,?)`, 3, 3, now, "Назначена", "", "Студент не явился на экзамен по дисциплине «Базы данных»")
	db.Exec(`INSERT INTO commissions (student_id, performance_id, assigned_date, status, decision, description)
		VALUES (?,?,?,?,?,?)`, 8, 13, now, "Проведена", "Допущен к пересдаче", "Не сдал лабораторные по ОС, допущен к пересдаче")
	db.Exec(`INSERT INTO commissions (student_id, performance_id, assigned_date, status, decision, description)
		VALUES (?,?,?,?,?,?)`, 14, 24, now, "Закрыта", "Условный перевод", "Перевод с условием пересдачи курсовой работы")

	log.Println("Demo data seeded successfully")
}

// --- Helpers ---

func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func parseJSON(r *http.Request, dest interface{}) error {
	return json.NewDecoder(r.Body).Decode(dest)
}

func getIDFromPath(r *http.Request) string {
	parts := strings.Split(r.URL.Path, "/")
	return parts[len(parts)-1]
}

// --- Login ---

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var creds struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := parseJSON(r, &creds); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	var id int
	var fio, position string
	err := db.QueryRow("SELECT id, fio, position FROM teachers WHERE login = ? AND password_hash = ?",
		creds.Login, hashPassword(creds.Password)).Scan(&id, &fio, &position)
	if err != nil {
		sendJSON(w, map[string]interface{}{"ok": false, "error": "Неверный логин или пароль"})
		return
	}
	sendJSON(w, map[string]interface{}{"ok": true, "id": id, "fio": fio, "position": position})
}

// --- Students CRUD ---

func getStudentsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, fio, record_book_number, group_name, course, study_form, training_direction, status FROM students ORDER BY group_name, fio")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var res []map[string]interface{}
	for rows.Next() {
		var id, course int
		var fio, rb, grp, form, dir, status string
		rows.Scan(&id, &fio, &rb, &grp, &course, &form, &dir, &status)
		res = append(res, map[string]interface{}{
			"id": id, "fio": fio, "record_book_number": rb, "group_name": grp,
			"course": course, "study_form": form, "training_direction": dir, "status": status,
		})
	}
	if res == nil {
		res = make([]map[string]interface{}, 0)
	}
	sendJSON(w, res)
}

func postStudentHandler(w http.ResponseWriter, r *http.Request) {
	var s map[string]interface{}
	parseJSON(r, &s)
	res, err := db.Exec("INSERT INTO students (fio, record_book_number, group_name, course, study_form, training_direction, status) VALUES (?,?,?,?,?,?,?)",
		s["fio"], s["record_book_number"], s["group_name"], s["course"], s["study_form"], s["training_direction"], s["status"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	sendJSON(w, map[string]interface{}{"id": id})
}

func putStudentHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r)
	var s map[string]interface{}
	parseJSON(r, &s)
	db.Exec("UPDATE students SET fio=?, record_book_number=?, group_name=?, course=?, study_form=?, training_direction=?, status=? WHERE id=?",
		s["fio"], s["record_book_number"], s["group_name"], s["course"], s["study_form"], s["training_direction"], s["status"], id)
	sendJSON(w, map[string]bool{"ok": true})
}

func deleteStudentHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r)
	db.Exec("DELETE FROM performance WHERE student_id=?", id)
	db.Exec("DELETE FROM commissions WHERE student_id=?", id)
	db.Exec("DELETE FROM students WHERE id=?", id)
	sendJSON(w, map[string]bool{"ok": true})
}

// --- Teachers CRUD ---

func getTeachersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, fio, position, academic_degree, login FROM teachers ORDER BY fio")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var res []map[string]interface{}
	for rows.Next() {
		var id int
		var fio, pos, deg, login string
		rows.Scan(&id, &fio, &pos, &deg, &login)
		res = append(res, map[string]interface{}{
			"id": id, "fio": fio, "position": pos, "academic_degree": deg, "login": login,
		})
	}
	if res == nil {
		res = make([]map[string]interface{}, 0)
	}
	sendJSON(w, res)
}

func postTeacherHandler(w http.ResponseWriter, r *http.Request) {
	var t map[string]interface{}
	parseJSON(r, &t)
	password := "123456"
	if p, ok := t["password"]; ok && p != nil && p != "" {
		password = p.(string)
	}
	res, err := db.Exec("INSERT INTO teachers (fio, position, academic_degree, login, password_hash) VALUES (?,?,?,?,?)",
		t["fio"], t["position"], t["academic_degree"], t["login"], hashPassword(password))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	sendJSON(w, map[string]interface{}{"id": id})
}

func putTeacherHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r)
	var t map[string]interface{}
	parseJSON(r, &t)
	db.Exec("UPDATE teachers SET fio=?, position=?, academic_degree=? WHERE id=?",
		t["fio"], t["position"], t["academic_degree"], id)
	sendJSON(w, map[string]bool{"ok": true})
}

func deleteTeacherHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r)
	db.Exec("DELETE FROM teachers WHERE id=?", id)
	sendJSON(w, map[string]bool{"ok": true})
}

// --- Disciplines CRUD ---

func getDisciplinesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT d.id, d.name, d.semester, d.control_type, d.hours, d.teacher_id,
		COALESCE(t.fio, '') as teacher_fio
		FROM disciplines d LEFT JOIN teachers t ON d.teacher_id = t.id ORDER BY d.semester, d.name`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var res []map[string]interface{}
	for rows.Next() {
		var id, semester, hours int
		var teacherId sql.NullInt64
		var name, controlType, teacherFio string
		rows.Scan(&id, &name, &semester, &controlType, &hours, &teacherId, &teacherFio)
		tid := 0
		if teacherId.Valid {
			tid = int(teacherId.Int64)
		}
		res = append(res, map[string]interface{}{
			"id": id, "name": name, "semester": semester, "control_type": controlType,
			"hours": hours, "teacher_id": tid, "teacher_fio": teacherFio,
		})
	}
	if res == nil {
		res = make([]map[string]interface{}, 0)
	}
	sendJSON(w, res)
}

func postDisciplineHandler(w http.ResponseWriter, r *http.Request) {
	var d map[string]interface{}
	parseJSON(r, &d)
	res, err := db.Exec("INSERT INTO disciplines (name, semester, control_type, hours, teacher_id) VALUES (?,?,?,?,?)",
		d["name"], d["semester"], d["control_type"], d["hours"], d["teacher_id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	sendJSON(w, map[string]interface{}{"id": id})
}

func putDisciplineHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r)
	var d map[string]interface{}
	parseJSON(r, &d)
	db.Exec("UPDATE disciplines SET name=?, semester=?, control_type=?, hours=?, teacher_id=? WHERE id=?",
		d["name"], d["semester"], d["control_type"], d["hours"], d["teacher_id"], id)
	sendJSON(w, map[string]bool{"ok": true})
}

func deleteDisciplineHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r)
	db.Exec("DELETE FROM disciplines WHERE id=?", id)
	sendJSON(w, map[string]bool{"ok": true})
}

// --- Performance CRUD ---

func getPerformanceHandler(w http.ResponseWriter, r *http.Request) {
	query := `SELECT p.id, p.student_id, p.discipline_id, p.teacher_id,
		p.sheet_number, p.sheet_fill_date, p.grade, p.status, p.comment, p.has_debt, p.created_at,
		COALESCE(s.fio,'') as student_fio, COALESCE(s.group_name,'') as student_group,
		COALESCE(d.name,'') as discipline_name, COALESCE(d.semester,0) as semester,
		COALESCE(t.fio,'') as teacher_fio
		FROM performance p
		LEFT JOIN students s ON p.student_id = s.id
		LEFT JOIN disciplines d ON p.discipline_id = d.id
		LEFT JOIN teachers t ON p.teacher_id = t.id`

	var conditions []string
	var args []interface{}

	q := r.URL.Query()
	if v := q.Get("group"); v != "" {
		conditions = append(conditions, "s.group_name = ?")
		args = append(args, v)
	}
	if v := q.Get("discipline_id"); v != "" {
		conditions = append(conditions, "p.discipline_id = ?")
		args = append(args, v)
	}
	if v := q.Get("semester"); v != "" {
		conditions = append(conditions, "d.semester = ?")
		args = append(args, v)
	}
	if v := q.Get("student_id"); v != "" {
		conditions = append(conditions, "p.student_id = ?")
		args = append(args, v)
	}
	if v := q.Get("has_debt"); v == "true" {
		conditions = append(conditions, "p.has_debt = 1")
	}
	if v := q.Get("sheet_number"); v != "" {
		conditions = append(conditions, "p.sheet_number = ?")
		args = append(args, v)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY s.group_name, s.fio, d.name"

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var res []map[string]interface{}
	for rows.Next() {
		var id, studentId, disciplineId, semester int
		var teacherId sql.NullInt64
		var hasDebt int
		var sheetNum, sheetDate, grade, status, comment, createdAt string
		var studentFio, studentGroup, discName, teacherFio string
		rows.Scan(&id, &studentId, &disciplineId, &teacherId,
			&sheetNum, &sheetDate, &grade, &status, &comment, &hasDebt, &createdAt,
			&studentFio, &studentGroup, &discName, &semester, &teacherFio)
		tid := 0
		if teacherId.Valid {
			tid = int(teacherId.Int64)
		}
		res = append(res, map[string]interface{}{
			"id": id, "student_id": studentId, "discipline_id": disciplineId, "teacher_id": tid,
			"sheet_number": sheetNum, "sheet_fill_date": sheetDate, "grade": grade,
			"status": status, "comment": comment, "has_debt": hasDebt == 1, "created_at": createdAt,
			"student_fio": studentFio, "student_group": studentGroup,
			"discipline_name": discName, "semester": semester, "teacher_fio": teacherFio,
		})
	}
	if res == nil {
		res = make([]map[string]interface{}, 0)
	}
	sendJSON(w, res)
}

func postPerformanceHandler(w http.ResponseWriter, r *http.Request) {
	var p map[string]interface{}
	parseJSON(r, &p)
	hasDebt := 0
	if v, ok := p["has_debt"]; ok && v == true {
		hasDebt = 1
	}
	now := time.Now().Format("2006-01-02")
	res, err := db.Exec(`INSERT INTO performance (student_id, discipline_id, teacher_id, sheet_number, sheet_fill_date, grade, status, comment, has_debt, created_at)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		p["student_id"], p["discipline_id"], p["teacher_id"], p["sheet_number"],
		p["sheet_fill_date"], p["grade"], p["status"], p["comment"], hasDebt, now)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	sendJSON(w, map[string]interface{}{"id": id})
}

func putPerformanceHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r)
	var p map[string]interface{}
	parseJSON(r, &p)
	hasDebt := 0
	if v, ok := p["has_debt"]; ok && v == true {
		hasDebt = 1
	}
	db.Exec(`UPDATE performance SET grade=?, status=?, comment=?, has_debt=? WHERE id=?`,
		p["grade"], p["status"], p["comment"], hasDebt, id)
	sendJSON(w, map[string]bool{"ok": true})
}

func deletePerformanceHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r)
	db.Exec("DELETE FROM commissions WHERE performance_id=?", id)
	db.Exec("DELETE FROM performance WHERE id=?", id)
	sendJSON(w, map[string]bool{"ok": true})
}

// Batch insert/update for sheet filling
func batchPerformanceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var batch struct {
		Records []map[string]interface{} `json:"records"`
	}
	if err := parseJSON(r, &batch); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	now := time.Now().Format("2006-01-02")
	for _, rec := range batch.Records {
		hasDebt := 0
		if v, ok := rec["has_debt"]; ok && v == true {
			hasDebt = 1
		}
		grade, _ := rec["grade"].(string)
		// Determine has_debt based on grade
		if grade == "Неудовлетворительно" || grade == "Не зачтено" {
			hasDebt = 1
		}
		status := "Сдано"
		if hasDebt == 1 {
			status = "Не сдано"
		}
		if s, ok := rec["status"].(string); ok && s != "" {
			status = s
		}

		// Check if record already exists
		var existingId int
		studentId := rec["student_id"]
		disciplineId := rec["discipline_id"]
		sheetNumber := rec["sheet_number"]

		err := db.QueryRow("SELECT id FROM performance WHERE student_id=? AND discipline_id=? AND sheet_number=?",
			studentId, disciplineId, sheetNumber).Scan(&existingId)

		comment := ""
		if c, ok := rec["comment"].(string); ok {
			comment = c
		}

		if err == nil {
			// Update existing
			db.Exec(`UPDATE performance SET grade=?, status=?, comment=?, has_debt=?, sheet_fill_date=? WHERE id=?`,
				grade, status, comment, hasDebt, now, existingId)
		} else {
			// Insert new
			teacherId := rec["teacher_id"]
			fillDate := now
			if fd, ok := rec["sheet_fill_date"].(string); ok && fd != "" {
				fillDate = fd
			}
			db.Exec(`INSERT INTO performance (student_id, discipline_id, teacher_id, sheet_number, sheet_fill_date, grade, status, comment, has_debt, created_at)
				VALUES (?,?,?,?,?,?,?,?,?,?)`,
				studentId, disciplineId, teacherId, sheetNumber, fillDate, grade, status, comment, hasDebt, now)
		}
	}
	sendJSON(w, map[string]bool{"ok": true})
}

// --- Monitoring ---

func getMonitoringHandler(w http.ResponseWriter, r *http.Request) {
	query := `SELECT p.id, p.student_id, p.discipline_id, p.teacher_id,
		p.sheet_number, p.grade, p.status, p.comment, p.has_debt,
		s.fio as student_fio, s.group_name, s.course,
		d.name as discipline_name, d.semester,
		COALESCE(t.fio,'') as teacher_fio
		FROM performance p
		JOIN students s ON p.student_id = s.id
		JOIN disciplines d ON p.discipline_id = d.id
		LEFT JOIN teachers t ON p.teacher_id = t.id
		WHERE (p.has_debt = 1 OR p.grade IN ('Неудовлетворительно', 'Не зачтено'))`

	var conditions []string
	var args []interface{}

	q := r.URL.Query()
	if v := q.Get("group"); v != "" {
		conditions = append(conditions, "s.group_name = ?")
		args = append(args, v)
	}
	if v := q.Get("discipline_id"); v != "" {
		conditions = append(conditions, "p.discipline_id = ?")
		args = append(args, v)
	}
	if v := q.Get("semester"); v != "" {
		conditions = append(conditions, "d.semester = ?")
		args = append(args, v)
	}
	if v := q.Get("status"); v != "" {
		conditions = append(conditions, "p.status = ?")
		args = append(args, v)
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY s.group_name, s.fio"

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var res []map[string]interface{}
	for rows.Next() {
		var id, studentId, disciplineId, course, semester int
		var teacherId sql.NullInt64
		var hasDebt int
		var sheetNum, grade, status, comment, studentFio, groupName, discName, teacherFio string
		rows.Scan(&id, &studentId, &disciplineId, &teacherId,
			&sheetNum, &grade, &status, &comment, &hasDebt,
			&studentFio, &groupName, &course, &discName, &semester, &teacherFio)
		tid := 0
		if teacherId.Valid {
			tid = int(teacherId.Int64)
		}

		// Check if already has commission
		var commissionExists bool
		db.QueryRow("SELECT EXISTS(SELECT 1 FROM commissions WHERE performance_id=?)", id).Scan(&commissionExists)

		res = append(res, map[string]interface{}{
			"performance_id": id, "student_id": studentId, "discipline_id": disciplineId, "teacher_id": tid,
			"sheet_number": sheetNum, "grade": grade, "status": status, "comment": comment,
			"has_debt": hasDebt == 1, "student_fio": studentFio, "group_name": groupName,
			"course": course, "discipline_name": discName, "semester": semester,
			"teacher_fio": teacherFio, "has_commission": commissionExists,
		})
	}
	if res == nil {
		res = make([]map[string]interface{}, 0)
	}

	// Calculate average grades per student
	avgRows, err := db.Query(`SELECT p.student_id, AVG(CASE
		WHEN p.grade = 'Отлично' THEN 5
		WHEN p.grade = 'Хорошо' THEN 4
		WHEN p.grade = 'Удовлетворительно' THEN 3
		WHEN p.grade = 'Неудовлетворительно' THEN 2
		WHEN p.grade = 'Зачтено' THEN 5
		WHEN p.grade = 'Не зачтено' THEN 2
		ELSE 0 END) as avg_grade
		FROM performance p GROUP BY p.student_id`)
	if err == nil {
		avgMap := make(map[int]float64)
		for avgRows.Next() {
			var sid int
			var avg float64
			avgRows.Scan(&sid, &avg)
			avgMap[sid] = avg
		}
		avgRows.Close()
		for i := range res {
			sid := int(res[i]["student_id"].(int))
			if avg, ok := avgMap[sid]; ok {
				res[i]["avg_grade"] = fmt.Sprintf("%.2f", avg)
			} else {
				res[i]["avg_grade"] = "—"
			}
		}
	}

	sendJSON(w, res)
}

// --- Commissions CRUD ---

func getCommissionsHandler(w http.ResponseWriter, r *http.Request) {
	query := `SELECT c.id, c.student_id, c.performance_id, c.assigned_date, c.status, c.decision, c.description,
		COALESCE(s.fio,'') as student_fio, COALESCE(s.group_name,'') as student_group,
		COALESCE(d.name,'') as discipline_name,
		COALESCE(p.grade,'') as grade
		FROM commissions c
		LEFT JOIN students s ON c.student_id = s.id
		LEFT JOIN performance p ON c.performance_id = p.id
		LEFT JOIN disciplines d ON p.discipline_id = d.id`

	var conditions []string
	var args []interface{}
	q := r.URL.Query()
	if v := q.Get("status"); v != "" {
		conditions = append(conditions, "c.status = ?")
		args = append(args, v)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY c.assigned_date DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var res []map[string]interface{}
	for rows.Next() {
		var id, studentId, perfId int
		var date, status, decision, desc, studentFio, studentGroup, discName, grade string
		rows.Scan(&id, &studentId, &perfId, &date, &status, &decision, &desc,
			&studentFio, &studentGroup, &discName, &grade)
		res = append(res, map[string]interface{}{
			"id": id, "student_id": studentId, "performance_id": perfId,
			"assigned_date": date, "status": status, "decision": decision, "description": desc,
			"student_fio": studentFio, "student_group": studentGroup,
			"discipline_name": discName, "grade": grade,
		})
	}
	if res == nil {
		res = make([]map[string]interface{}, 0)
	}
	sendJSON(w, res)
}

func postCommissionHandler(w http.ResponseWriter, r *http.Request) {
	var c map[string]interface{}
	parseJSON(r, &c)
	now := time.Now().Format("2006-01-02")
	date := now
	if d, ok := c["assigned_date"].(string); ok && d != "" {
		date = d
	}
	res, err := db.Exec(`INSERT INTO commissions (student_id, performance_id, assigned_date, status, decision, description)
		VALUES (?,?,?,?,?,?)`,
		c["student_id"], c["performance_id"], date, "Назначена", "", c["description"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	sendJSON(w, map[string]interface{}{"id": id})
}

func putCommissionHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r)
	var c map[string]interface{}
	parseJSON(r, &c)
	if status, ok := c["status"]; ok {
		db.Exec("UPDATE commissions SET status=? WHERE id=?", status, id)
	}
	if decision, ok := c["decision"]; ok {
		db.Exec("UPDATE commissions SET decision=? WHERE id=?", decision, id)
	}
	if desc, ok := c["description"]; ok {
		db.Exec("UPDATE commissions SET description=? WHERE id=?", desc, id)
	}
	sendJSON(w, map[string]bool{"ok": true})
}

// --- Groups list helper ---

func getGroupsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT DISTINCT group_name FROM students ORDER BY group_name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var groups []string
	for rows.Next() {
		var g string
		rows.Scan(&g)
		groups = append(groups, g)
	}
	if groups == nil {
		groups = make([]string, 0)
	}
	sendJSON(w, groups)
}

// --- Semesters list helper ---

func getSemestersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT DISTINCT semester FROM disciplines ORDER BY semester")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var semesters []int
	for rows.Next() {
		var s int
		rows.Scan(&s)
		semesters = append(semesters, s)
	}
	if semesters == nil {
		semesters = make([]int, 0)
	}
	sendJSON(w, semesters)
}

// --- Students by group helper ---

func getStudentsByGroupHandler(w http.ResponseWriter, r *http.Request) {
	group := r.URL.Query().Get("group")
	if group == "" {
		sendJSON(w, []map[string]interface{}{})
		return
	}
	rows, err := db.Query("SELECT id, fio, record_book_number FROM students WHERE group_name = ? AND status = 'Активен' ORDER BY fio", group)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var res []map[string]interface{}
	for rows.Next() {
		var id int
		var fio, rb string
		rows.Scan(&id, &fio, &rb)
		res = append(res, map[string]interface{}{"id": id, "fio": fio, "record_book_number": rb})
	}
	if res == nil {
		res = make([]map[string]interface{}, 0)
	}
	sendJSON(w, res)
}

// --- Router helper ---

func routeHandler(get, post, put, del http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if get != nil {
				get(w, r)
			} else {
				sendJSON(w, []struct{}{})
			}
		case http.MethodPost:
			if post != nil {
				post(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		case http.MethodPut:
			if put != nil {
				put(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		case http.MethodDelete:
			if del != nil {
				del(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

// --- Stats endpoint ---

func getStatsHandler(w http.ResponseWriter, r *http.Request) {
	var totalStudents, totalTeachers, totalDisciplines, totalDebts, totalCommissions int
	db.QueryRow("SELECT COUNT(*) FROM students WHERE status='Активен'").Scan(&totalStudents)
	db.QueryRow("SELECT COUNT(*) FROM teachers").Scan(&totalTeachers)
	db.QueryRow("SELECT COUNT(*) FROM disciplines").Scan(&totalDisciplines)
	db.QueryRow("SELECT COUNT(*) FROM performance WHERE has_debt=1").Scan(&totalDebts)
	db.QueryRow("SELECT COUNT(*) FROM commissions WHERE status != 'Закрыта'").Scan(&totalCommissions)

	var avgGrade float64
	db.QueryRow(`SELECT COALESCE(AVG(CASE
		WHEN grade = 'Отлично' THEN 5
		WHEN grade = 'Хорошо' THEN 4
		WHEN grade = 'Удовлетворительно' THEN 3
		WHEN grade = 'Неудовлетворительно' THEN 2
		WHEN grade = 'Зачтено' THEN 5
		WHEN grade = 'Не зачтено' THEN 2
		ELSE 0 END), 0) FROM performance`).Scan(&avgGrade)

	sendJSON(w, map[string]interface{}{
		"total_students":    totalStudents,
		"total_teachers":    totalTeachers,
		"total_disciplines": totalDisciplines,
		"total_debts":       totalDebts,
		"total_commissions": totalCommissions,
		"avg_grade":         strconv.FormatFloat(avgGrade, 'f', 2, 64),
	})
}

func main() {
	err := initDB(filepath.Join(".", "database.sqlite"))
	if err != nil {
		log.Fatalf("Failed to init db: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("/api/login", loginHandler)

	// Stats
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getStatsHandler(w, r)
		}
	})

	// Students
	mux.HandleFunc("/api/students", routeHandler(getStudentsHandler, postStudentHandler, nil, nil))
	mux.HandleFunc("/api/students/", routeHandler(nil, nil, putStudentHandler, deleteStudentHandler))

	// Teachers
	mux.HandleFunc("/api/teachers", routeHandler(getTeachersHandler, postTeacherHandler, nil, nil))
	mux.HandleFunc("/api/teachers/", routeHandler(nil, nil, putTeacherHandler, deleteTeacherHandler))

	// Disciplines
	mux.HandleFunc("/api/disciplines", routeHandler(getDisciplinesHandler, postDisciplineHandler, nil, nil))
	mux.HandleFunc("/api/disciplines/", routeHandler(nil, nil, putDisciplineHandler, deleteDisciplineHandler))

	// Performance
	mux.HandleFunc("/api/performance", routeHandler(getPerformanceHandler, postPerformanceHandler, nil, nil))
	mux.HandleFunc("/api/performance/", routeHandler(nil, nil, putPerformanceHandler, deletePerformanceHandler))
	mux.HandleFunc("/api/performance/batch", batchPerformanceHandler)

	// Monitoring
	mux.HandleFunc("/api/monitoring", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getMonitoringHandler(w, r)
		}
	})

	// Commissions
	mux.HandleFunc("/api/commissions", routeHandler(getCommissionsHandler, postCommissionHandler, nil, nil))
	mux.HandleFunc("/api/commissions/", routeHandler(nil, nil, putCommissionHandler, nil))

	// Helper endpoints
	mux.HandleFunc("/api/groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getGroupsHandler(w, r)
		}
	})
	mux.HandleFunc("/api/semesters", func(w http.ResponseWriter, r *http.Request) { //
		if r.Method == http.MethodGet {
			getSemestersHandler(w, r)
		}
	})
	mux.HandleFunc("/api/students-by-group", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getStudentsByGroupHandler(w, r)
		}
	})

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
	})

	srv := &http.Server{Addr: ":3000", Handler: c.Handler(mux)}
	log.Println("API Server running on http://localhost:3000")
	srv.ListenAndServe()
}
