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

// ==================== DB INIT ====================

func initDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		return err
	}

	tables := []string{
		// --- Main tables ---
		`CREATE TABLE IF NOT EXISTS groups_t (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			course INTEGER NOT NULL DEFAULT 1,
			study_form TEXT NOT NULL DEFAULT '',
			training_direction TEXT NOT NULL DEFAULT ''
		);`,
		`CREATE TABLE IF NOT EXISTS students (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			fio TEXT NOT NULL,
			record_book_number TEXT NOT NULL UNIQUE,
			group_id INTEGER NOT NULL,
			status TEXT NOT NULL DEFAULT 'Активен',
			FOREIGN KEY(group_id) REFERENCES groups_t(id)
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
			control_type TEXT NOT NULL DEFAULT '',
			hours INTEGER NOT NULL DEFAULT 0,
			teacher_id INTEGER,
			FOREIGN KEY(teacher_id) REFERENCES teachers(id)
		);`,
		`CREATE TABLE IF NOT EXISTS sheets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			number TEXT NOT NULL UNIQUE,
			fill_date TEXT NOT NULL DEFAULT '',
			group_id INTEGER NOT NULL,
			discipline_id INTEGER NOT NULL,
			FOREIGN KEY(group_id) REFERENCES groups_t(id),
			FOREIGN KEY(discipline_id) REFERENCES disciplines(id)
		);`,
		`CREATE TABLE IF NOT EXISTS performance (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sheet_id INTEGER NOT NULL,
			student_id INTEGER NOT NULL,
			grade TEXT NOT NULL DEFAULT '',
			comment TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT '',
			has_debt INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT '',
			FOREIGN KEY(sheet_id) REFERENCES sheets(id),
			FOREIGN KEY(student_id) REFERENCES students(id)
		);`,
		`CREATE TABLE IF NOT EXISTS commissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			performance_id INTEGER NOT NULL,
			assigned_date TEXT NOT NULL DEFAULT '',
			meeting_date TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'Назначена',
			decision TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			FOREIGN KEY(performance_id) REFERENCES performance(id)
		);`,
		`CREATE TABLE IF NOT EXISTS commission_members (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			commission_id INTEGER NOT NULL,
			teacher_id INTEGER NOT NULL,
			role TEXT NOT NULL DEFAULT 'Член комиссии',
			FOREIGN KEY(commission_id) REFERENCES commissions(id),
			FOREIGN KEY(teacher_id) REFERENCES teachers(id)
		);`,
	}

	for _, query := range tables {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("create table: %w", err)
		}
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM teachers").Scan(&count)
	if count == 0 {
		seedData()
	}
	return nil
}

func hashPassword(p string) string {
	h := sha256.Sum256([]byte(p))
	return fmt.Sprintf("%x", h)
}

// ==================== SEED ====================

func seedData() {
	now := time.Now().Format("2006-01-02")

	// Groups: name, course, study_form, training_direction
	groups := []struct {
		name string; course int; form, dir string
	}{
		{"ИС-21", 2, "Очная", "Информационные системы"}, 
		{"ПИ-31", 3, "Очная", "Программная инженерия"}, 
		{"КБ-41", 4, "Очная", "Кибербезопасность"},
	}
	for _, g := range groups {
		db.Exec("INSERT INTO groups_t(name,course,study_form,training_direction) VALUES(?,?,?,?)", g.name, g.course, g.form, g.dir)
	}

	// Teachers
	teachers := []struct {
		fio, pos, degree, login string
	}{
		{"Иванов Сергей Петрович", "Профессор", "Доктор наук", "ivanov"},
		{"Петрова Елена Алексеевна", "Доцент", "Кандидат наук", "petrova"},
		{"Сидоров Дмитрий Николаевич", "Старший преподаватель", "Кандидат наук", "sidorov"},
		{"Козлова Анна Михайловна", "Доцент", "Кандидат наук", "kozlova"},
		{"Новиков Алексей Владимирович", "Профессор", "Доктор наук", "novikov"},
	}
	for _, t := range teachers {
		db.Exec("INSERT INTO teachers(fio,position,academic_degree,login,password_hash) VALUES(?,?,?,?,?)",
			t.fio, t.pos, t.degree, t.login, hashPassword("123456"))
	}

	// Students
	studs := []struct {
		fio, rb string; gid int
	}{
		{"Абрамов Иван Сергеевич", "2024-001", 1},
		{"Белова Мария Дмитриевна", "2024-002", 1},
		{"Васильев Пётр Андреевич", "2024-003", 1},
		{"Григорьева Анна Олеговна", "2024-004", 1},
		{"Дмитриев Олег Игоревич", "2024-005", 1},
		{"Егоров Максим Викторович", "2024-006", 2},
		{"Жукова Елена Сергеевна", "2024-007", 2},
		{"Зайцев Артём Павлович", "2024-008", 2},
		{"Ильина Дарья Романовна", "2024-009", 2},
		{"Кузнецов Никита Алексеевич", "2024-010", 2},
		{"Лебедев Роман Дмитриевич", "2024-011", 3},
		{"Морозова Ольга Ивановна", "2024-012", 3},
		{"Николаев Сергей Владимирович", "2024-013", 3},
		{"Орлова Виктория Андреевна", "2024-014", 3},
		{"Павлов Денис Олегович", "2024-015", 3},
	}
	for _, s := range studs {
		db.Exec("INSERT INTO students(fio,record_book_number,group_id,status) VALUES(?,?,?,?)", s.fio, s.rb, s.gid, "Активен")
	}

	// Disciplines: name, semester, control_type, hours, teacher_id
	discs := []struct {
		name string; sem int; ctrl string; hrs, tid int
	}{
		{"Базы данных", 3, "Экзамен", 108, 1},
		{"Программирование на Go", 4, "Экзамен", 72, 1},
		{"Математический анализ", 3, "Экзамен", 144, 2},
		{"Операционные системы", 5, "Зачёт", 90, 3},
		{"Компьютерные сети", 6, "Экзамен", 108, 4},
		{"Информационная безопасность", 7, "Зачёт", 72, 5},
	}
	for _, d := range discs {
		db.Exec("INSERT INTO disciplines(name,semester,control_type,hours,teacher_id) VALUES(?,?,?,?,?)",
			d.name, d.sem, d.ctrl, d.hrs, d.tid)
	}

	// Sheets: number, fill_date, group_id, discipline_id
	shts := []struct {
		num string; gid, did int
	}{
		{"ВД-001", 1, 1}, {"ВД-002", 1, 3}, {"ВД-003", 2, 4},
		{"ВД-004", 2, 2}, {"ВД-005", 3, 5}, {"ВД-006", 3, 6},
	}
	for _, s := range shts {
		db.Exec("INSERT INTO sheets(number,fill_date,group_id,discipline_id) VALUES(?,?,?,?)", s.num, now, s.gid, s.did)
	}

	// Performance
	type pRec struct {
		sheetID, studentID int
		grade, status, comment string
		hasDebt int
	}
	perfs := []pRec{
		{1,1,"Отлично","Сдано","",0},{1,2,"Хорошо","Сдано","",0},
		{1,3,"Неудовлетворительно","Не сдано","Не явился на экзамен",1},
		{1,4,"Удовлетворительно","Сдано","",0},
		{1,5,"Неудовлетворительно","Не сдано","Плохая подготовка",1},
		{2,1,"Хорошо","Сдано","",0},{2,2,"Отлично","Сдано","",0},
		{2,3,"Неудовлетворительно","Не сдано","Систематические пропуски",1},
		{2,4,"Хорошо","Сдано","",0},{2,5,"Удовлетворительно","Сдано","",0},
		{3,6,"Зачтено","Сдано","",0},{3,7,"Зачтено","Сдано","",0},
		{3,8,"Не зачтено","Не сдано","Не сдал лабораторные работы",1},
		{3,9,"Зачтено","Сдано","",0},
		{3,10,"Не зачтено","Не сдано","Низкая посещаемость",1},
		{4,6,"Отлично","Сдано","",0},{4,7,"Хорошо","Сдано","",0},
		{4,8,"Удовлетворительно","Сдано","",0},
		{4,9,"Неудовлетворительно","Не сдано","Не сдал проект",1},
		{4,10,"Хорошо","Сдано","",0},
		{5,11,"Отлично","Сдано","",0},{5,12,"Хорошо","Сдано","",0},
		{5,13,"Удовлетворительно","Сдано","",0},
		{5,14,"Неудовлетворительно","Не сдано","Не выполнил курсовую работу",1},
		{5,15,"Удовлетворительно","Сдано","",0},
		{6,11,"Зачтено","Сдано","",0},
		{6,12,"Не зачтено","Не сдано","Не защитил лабораторные",1},
		{6,13,"Зачтено","Сдано","",0},
		{6,14,"Не зачтено","Не сдано","Пропуски занятий",1},
		{6,15,"Зачтено","Сдано","",0},
	}
	for _, p := range perfs {
		db.Exec(`INSERT INTO performance(sheet_id,student_id,grade,comment,status,has_debt,created_at) VALUES(?,?,?,?,?,?,?)`,
			p.sheetID, p.studentID, p.grade, p.comment, p.status, p.hasDebt, now)
	}

	// Commissions
	db.Exec(`INSERT INTO commissions(performance_id,assigned_date,meeting_date,status,decision,description) VALUES(?,?,?,?,?,?)`,
		3, now, "", "Назначена", "", "Студент не явился на экзамен по дисциплине «Базы данных»")
	db.Exec(`INSERT INTO commissions(performance_id,assigned_date,meeting_date,status,decision,description) VALUES(?,?,?,?,?,?)`,
		13, now, now, "Проведена", "Допущен к пересдаче", "Не сдал лабораторные по ОС")
	db.Exec(`INSERT INTO commissions(performance_id,assigned_date,meeting_date,status,decision,description) VALUES(?,?,?,?,?,?)`,
		24, now, now, "Закрыта", "Условный перевод", "Перевод с условием пересдачи курсовой работы")

	// Commission members
	db.Exec("INSERT INTO commission_members(commission_id,teacher_id,role) VALUES(?,?,?)", 1, 1, "Член комиссии")
	db.Exec("INSERT INTO commission_members(commission_id,teacher_id,role) VALUES(?,?,?)", 1, 2, "Член комиссии")
	db.Exec("INSERT INTO commission_members(commission_id,teacher_id,role) VALUES(?,?,?)", 1, 3, "Член комиссии")
	db.Exec("INSERT INTO commission_members(commission_id,teacher_id,role) VALUES(?,?,?)", 2, 3, "Член комиссии")
	db.Exec("INSERT INTO commission_members(commission_id,teacher_id,role) VALUES(?,?,?)", 2, 1, "Член комиссии")
	db.Exec("INSERT INTO commission_members(commission_id,teacher_id,role) VALUES(?,?,?)", 3, 4, "Член комиссии")
	db.Exec("INSERT INTO commission_members(commission_id,teacher_id,role) VALUES(?,?,?)", 3, 5, "Член комиссии")
	db.Exec("INSERT INTO commission_members(commission_id,teacher_id,role) VALUES(?,?,?)", 3, 1, "Член комиссии")

	log.Println("Demo data seeded successfully")
}

// ==================== HELPERS ====================

func sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func parseJSON(r *http.Request, dest interface{}) error {
	return json.NewDecoder(r.Body).Decode(dest)
}

func idFromPath(r *http.Request) string {
	parts := strings.Split(r.URL.Path, "/")
	return parts[len(parts)-1]
}

func emptyArr() []map[string]interface{} {
	return make([]map[string]interface{}, 0)
}

// ==================== LOGIN ====================

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w, "method", 405); return }
	var c struct{ Login, Password string }
	parseJSON(r, &c)
	var id int; var fio, pos string
	err := db.QueryRow(`SELECT id, fio, position FROM teachers WHERE login=? AND password_hash=?`,
		c.Login, hashPassword(c.Password)).Scan(&id, &fio, &pos)
	if err != nil { sendJSON(w, map[string]interface{}{"ok": false, "error": "Неверный логин или пароль"}); return }
	sendJSON(w, map[string]interface{}{"ok": true, "id": id, "fio": fio, "position": pos})
}

// ==================== GROUPS ====================

func getGroupsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT id, name, course, training_direction, study_form FROM groups_t ORDER BY name`)
	if err != nil { http.Error(w, err.Error(), 500); return }
	defer rows.Close()
	var res []map[string]interface{}
	for rows.Next() {
		var id, course int; var name, dir, form string
		rows.Scan(&id, &name, &course, &dir, &form)
		res = append(res, map[string]interface{}{
			"id": id, "name": name, "course": course, "training_direction": dir, "study_form": form,
		})
	}
	if res == nil { res = emptyArr() }
	sendJSON(w, res)
}

// ==================== STUDENTS ====================

func getStudentsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT s.id, s.fio, s.record_book_number, s.group_id, s.status,
		g.name, g.course, g.training_direction, g.study_form
		FROM students s JOIN groups_t g ON s.group_id=g.id
		ORDER BY g.name, s.fio`)
	if err != nil { http.Error(w, err.Error(), 500); return }
	defer rows.Close()
	var res []map[string]interface{}
	for rows.Next() {
		var id, gid, course int; var fio, rb, status, gname, dir, form string
		rows.Scan(&id, &fio, &rb, &gid, &status, &gname, &course, &dir, &form)
		res = append(res, map[string]interface{}{
			"id": id, "fio": fio, "record_book_number": rb, "group_id": gid, "status": status,
			"group_name": gname, "course": course, "training_direction": dir, "study_form": form,
		})
	}
	if res == nil { res = emptyArr() }
	sendJSON(w, res)
}

func postStudentHandler(w http.ResponseWriter, r *http.Request) {
	var s map[string]interface{}; parseJSON(r, &s)
	res, err := db.Exec("INSERT INTO students(fio,record_book_number,group_id,status) VALUES(?,?,?,?)",
		s["fio"], s["record_book_number"], s["group_id"], s["status"])
	if err != nil { http.Error(w, err.Error(), 500); return }
	id, _ := res.LastInsertId()
	sendJSON(w, map[string]interface{}{"id": id})
}

func putStudentHandler(w http.ResponseWriter, r *http.Request) {
	sid := idFromPath(r); var s map[string]interface{}; parseJSON(r, &s)
	db.Exec("UPDATE students SET fio=?,record_book_number=?,group_id=?,status=? WHERE id=?",
		s["fio"], s["record_book_number"], s["group_id"], s["status"], sid)
	sendJSON(w, map[string]bool{"ok": true})
}

func deleteStudentHandler(w http.ResponseWriter, r *http.Request) {
	sid := idFromPath(r)
	db.Exec("DELETE FROM commission_members WHERE commission_id IN (SELECT c.id FROM commissions c JOIN performance p ON c.performance_id=p.id WHERE p.student_id=?)", sid)
	db.Exec("DELETE FROM commissions WHERE performance_id IN (SELECT id FROM performance WHERE student_id=?)", sid)
	db.Exec("DELETE FROM performance WHERE student_id=?", sid)
	db.Exec("DELETE FROM students WHERE id=?", sid)
	sendJSON(w, map[string]bool{"ok": true})
}

// ==================== TEACHERS ====================

func getTeachersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT id, fio, position, academic_degree, login FROM teachers ORDER BY fio`)
	if err != nil { http.Error(w, err.Error(), 500); return }
	defer rows.Close()
	var res []map[string]interface{}
	for rows.Next() {
		var id int; var fio, pos, deg, login string
		rows.Scan(&id, &fio, &pos, &deg, &login)
		res = append(res, map[string]interface{}{
			"id": id, "fio": fio, "position": pos,
			"academic_degree": deg, "login": login,
		})
	}
	if res == nil { res = emptyArr() }
	sendJSON(w, res)
}

func postTeacherHandler(w http.ResponseWriter, r *http.Request) {
	var t map[string]interface{}; parseJSON(r, &t)
	pw := "123456"; if p, ok := t["password"]; ok && p != nil && p != "" { pw = p.(string) }
	res, err := db.Exec("INSERT INTO teachers(fio,position,academic_degree,login,password_hash) VALUES(?,?,?,?,?)",
		t["fio"], t["position"], t["academic_degree"], t["login"], hashPassword(pw))
	if err != nil { http.Error(w, err.Error(), 500); return }
	id, _ := res.LastInsertId()
	sendJSON(w, map[string]interface{}{"id": id})
}

func putTeacherHandler(w http.ResponseWriter, r *http.Request) {
	tid := idFromPath(r); var t map[string]interface{}; parseJSON(r, &t)
	db.Exec("UPDATE teachers SET fio=?,position=?,academic_degree=? WHERE id=?",
		t["fio"], t["position"], t["academic_degree"], tid)
	sendJSON(w, map[string]bool{"ok": true})
}

func deleteTeacherHandler(w http.ResponseWriter, r *http.Request) {
	tid := idFromPath(r)
	db.Exec("DELETE FROM commission_members WHERE teacher_id=?", tid)
	db.Exec("DELETE FROM teachers WHERE id=?", tid)
	sendJSON(w, map[string]bool{"ok": true})
}

// ==================== DISCIPLINES ====================

func getDisciplinesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT d.id, d.name, d.semester, d.hours, d.teacher_id, d.control_type,
		COALESCE(t.fio,'')
		FROM disciplines d LEFT JOIN teachers t ON d.teacher_id=t.id ORDER BY d.semester, d.name`)
	if err != nil { http.Error(w, err.Error(), 500); return }
	defer rows.Close()
	var res []map[string]interface{}
	for rows.Next() {
		var id, sem, hrs int; var tid sql.NullInt64
		var name, tFio, ctName string
		rows.Scan(&id, &name, &sem, &hrs, &tid, &ctName, &tFio)
		t := 0; if tid.Valid { t = int(tid.Int64) }
		res = append(res, map[string]interface{}{
			"id": id, "name": name, "semester": sem, "hours": hrs,
			"teacher_id": t, "control_type": ctName, "teacher_fio": tFio,
		})
	}
	if res == nil { res = emptyArr() }
	sendJSON(w, res)
}

func postDisciplineHandler(w http.ResponseWriter, r *http.Request) {
	var d map[string]interface{}; parseJSON(r, &d)
	res, err := db.Exec("INSERT INTO disciplines(name,semester,hours,teacher_id,control_type) VALUES(?,?,?,?,?)",
		d["name"], d["semester"], d["hours"], d["teacher_id"], d["control_type"])
	if err != nil { http.Error(w, err.Error(), 500); return }
	id, _ := res.LastInsertId()
	sendJSON(w, map[string]interface{}{"id": id})
}

func putDisciplineHandler(w http.ResponseWriter, r *http.Request) {
	did := idFromPath(r); var d map[string]interface{}; parseJSON(r, &d)
	db.Exec("UPDATE disciplines SET name=?,semester=?,hours=?,teacher_id=?,control_type=? WHERE id=?",
		d["name"], d["semester"], d["hours"], d["teacher_id"], d["control_type"], did)
	sendJSON(w, map[string]bool{"ok": true})
}

func deleteDisciplineHandler(w http.ResponseWriter, r *http.Request) {
	did := idFromPath(r)
	db.Exec("DELETE FROM disciplines WHERE id=?", did)
	sendJSON(w, map[string]bool{"ok": true})
}

// ==================== SHEETS ====================

func getSheetsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT sh.id, sh.number, sh.fill_date, sh.group_id, sh.discipline_id,
		g.name, d.name FROM sheets sh
		JOIN groups_t g ON sh.group_id=g.id JOIN disciplines d ON sh.discipline_id=d.id
		ORDER BY sh.number`)
	if err != nil { http.Error(w, err.Error(), 500); return }
	defer rows.Close()
	var res []map[string]interface{}
	for rows.Next() {
		var id, gid, did int; var num, date, gname, dname string
		rows.Scan(&id, &num, &date, &gid, &did, &gname, &dname)
		res = append(res, map[string]interface{}{
			"id": id, "number": num, "fill_date": date, "group_id": gid,
			"discipline_id": did, "group_name": gname, "discipline_name": dname,
		})
	}
	if res == nil { res = emptyArr() }
	sendJSON(w, res)
}

func postSheetHandler(w http.ResponseWriter, r *http.Request) {
	var s map[string]interface{}; parseJSON(r, &s)
	now := time.Now().Format("2006-01-02")
	fd := now; if d, ok := s["fill_date"].(string); ok && d != "" { fd = d }
	// Return existing if same number
	var existID int
	err := db.QueryRow("SELECT id FROM sheets WHERE number=?", s["number"]).Scan(&existID)
	if err == nil { sendJSON(w, map[string]interface{}{"id": existID}); return }
	res, err2 := db.Exec("INSERT INTO sheets(number,fill_date,group_id,discipline_id) VALUES(?,?,?,?)",
		s["number"], fd, s["group_id"], s["discipline_id"])
	if err2 != nil { http.Error(w, err2.Error(), 500); return }
	id, _ := res.LastInsertId()
	sendJSON(w, map[string]interface{}{"id": id})
}

// ==================== PERFORMANCE ====================

func getPerformanceHandler(w http.ResponseWriter, r *http.Request) {
	q := `SELECT p.id, p.sheet_id, p.student_id, p.grade, p.comment, p.status, p.has_debt, p.created_at,
		s.fio, g.name, sh.number, d.name, d.semester, COALESCE(t.fio,'')
		FROM performance p
		JOIN students s ON p.student_id=s.id JOIN groups_t g ON s.group_id=g.id
		JOIN sheets sh ON p.sheet_id=sh.id JOIN disciplines d ON sh.discipline_id=d.id
		LEFT JOIN teachers t ON d.teacher_id=t.id`

	var conds []string; var args []interface{}
	qp := r.URL.Query()
	if v := qp.Get("group_id"); v != "" { conds = append(conds, "s.group_id=?"); args = append(args, v) }
	if v := qp.Get("discipline_id"); v != "" { conds = append(conds, "sh.discipline_id=?"); args = append(args, v) }
	if v := qp.Get("semester"); v != "" { conds = append(conds, "d.semester=?"); args = append(args, v) }
	if v := qp.Get("has_debt"); v == "true" { conds = append(conds, "p.has_debt=1") }
	if v := qp.Get("sheet_id"); v != "" { conds = append(conds, "p.sheet_id=?"); args = append(args, v) }
	if len(conds) > 0 { q += " WHERE " + strings.Join(conds, " AND ") }
	q += " ORDER BY g.name, s.fio, d.name"

	rows, err := db.Query(q, args...)
	if err != nil { http.Error(w, err.Error(), 500); return }
	defer rows.Close()
	var res []map[string]interface{}
	for rows.Next() {
		var id, sheetID, studentID, sem, hasDebt int
		var grade, comment, status, createdAt, sFio, gName, shNum, dName, tFio string
		rows.Scan(&id, &sheetID, &studentID, &grade, &comment, &status, &hasDebt, &createdAt,
			&sFio, &gName, &shNum, &dName, &sem, &tFio)
		res = append(res, map[string]interface{}{
			"id": id, "sheet_id": sheetID, "student_id": studentID, "grade": grade,
			"comment": comment, "status": status, "has_debt": hasDebt == 1, "created_at": createdAt,
			"student_fio": sFio, "group_name": gName, "sheet_number": shNum,
			"discipline_name": dName, "semester": sem, "teacher_fio": tFio,
		})
	}
	if res == nil { res = emptyArr() }
	sendJSON(w, res)
}

func postPerformanceHandler(w http.ResponseWriter, r *http.Request) {
	var p map[string]interface{}; parseJSON(r, &p)
	hd := 0; if v, ok := p["has_debt"]; ok && v == true { hd = 1 }
	now := time.Now().Format("2006-01-02")
	res, err := db.Exec(`INSERT INTO performance(sheet_id,student_id,grade,comment,status,has_debt,created_at) VALUES(?,?,?,?,?,?,?)`,
		p["sheet_id"], p["student_id"], p["grade"], p["comment"], p["status"], hd, now)
	if err != nil { http.Error(w, err.Error(), 500); return }
	id, _ := res.LastInsertId()
	sendJSON(w, map[string]interface{}{"id": id})
}

func putPerformanceHandler(w http.ResponseWriter, r *http.Request) {
	pid := idFromPath(r); var p map[string]interface{}; parseJSON(r, &p)
	hd := 0; if v, ok := p["has_debt"]; ok && v == true { hd = 1 }
	db.Exec("UPDATE performance SET grade=?,status=?,comment=?,has_debt=? WHERE id=?",
		p["grade"], p["status"], p["comment"], hd, pid)
	sendJSON(w, map[string]bool{"ok": true})
}

func deletePerformanceHandler(w http.ResponseWriter, r *http.Request) {
	pid := idFromPath(r)
	db.Exec("DELETE FROM commission_members WHERE commission_id IN (SELECT id FROM commissions WHERE performance_id=?)", pid)
	db.Exec("DELETE FROM commissions WHERE performance_id=?", pid)
	db.Exec("DELETE FROM performance WHERE id=?", pid)
	sendJSON(w, map[string]bool{"ok": true})
}

// Batch performance for sheet filling
func batchPerformanceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w, "method", 405); return }
	var batch struct{ Records []map[string]interface{} `json:"records"` }
	if err := parseJSON(r, &batch); err != nil { http.Error(w, "bad request", 400); return }
	now := time.Now().Format("2006-01-02")
	for _, rec := range batch.Records {
		hd := 0
		grade, _ := rec["grade"].(string)
		if grade == "Неудовлетворительно" || grade == "Не зачтено" { hd = 1 }
		st := "Сдано"; if hd == 1 { st = "Не сдано" }
		if s, ok := rec["status"].(string); ok && s != "" { st = s }
		comment := ""; if c, ok := rec["comment"].(string); ok { comment = c }
		sheetID := rec["sheet_id"]; studentID := rec["student_id"]
		var existID int
		err := db.QueryRow("SELECT id FROM performance WHERE sheet_id=? AND student_id=?", sheetID, studentID).Scan(&existID)
		if err == nil {
			db.Exec("UPDATE performance SET grade=?,status=?,comment=?,has_debt=?,created_at=? WHERE id=?",
				grade, st, comment, hd, now, existID)
		} else {
			db.Exec("INSERT INTO performance(sheet_id,student_id,grade,comment,status,has_debt,created_at) VALUES(?,?,?,?,?,?,?)",
				sheetID, studentID, grade, comment, st, hd, now)
		}
	}
	sendJSON(w, map[string]bool{"ok": true})
}

// ==================== MONITORING ====================

func getMonitoringHandler(w http.ResponseWriter, r *http.Request) {
	q := `SELECT p.id, p.student_id, p.grade, p.status, p.comment, p.has_debt,
		s.fio, g.name, g.course, d.id, d.name, d.semester, COALESCE(t.id,0), COALESCE(t.fio,'')
		FROM performance p
		JOIN students s ON p.student_id=s.id JOIN groups_t g ON s.group_id=g.id
		JOIN sheets sh ON p.sheet_id=sh.id JOIN disciplines d ON sh.discipline_id=d.id
		LEFT JOIN teachers t ON d.teacher_id=t.id
		WHERE (p.has_debt=1 OR p.grade IN ('Неудовлетворительно','Не зачтено'))`

	var conds []string; var args []interface{}
	qp := r.URL.Query()
	if v := qp.Get("group_id"); v != "" { conds = append(conds, "s.group_id=?"); args = append(args, v) }
	if v := qp.Get("discipline_id"); v != "" { conds = append(conds, "sh.discipline_id=?"); args = append(args, v) }
	if v := qp.Get("semester"); v != "" { conds = append(conds, "d.semester=?"); args = append(args, v) }
	if v := qp.Get("teacher_id"); v != "" { conds = append(conds, "d.teacher_id=?"); args = append(args, v) }
	if len(conds) > 0 { q += " AND " + strings.Join(conds, " AND ") }
	q += " ORDER BY g.name, s.fio"

	rows, err := db.Query(q, args...)
	if err != nil { http.Error(w, err.Error(), 500); return }
	defer rows.Close()

	var res []map[string]interface{}
	for rows.Next() {
		var perfID, studentID, course, discID, sem, teacherID, hasDebt int
		var grade, status, comment, sFio, gName, dName, tFio string
		rows.Scan(&perfID, &studentID, &grade, &status, &comment, &hasDebt,
			&sFio, &gName, &course, &discID, &dName, &sem, &teacherID, &tFio)

		var commExists bool
		db.QueryRow("SELECT EXISTS(SELECT 1 FROM commissions WHERE performance_id=?)", perfID).Scan(&commExists)

		res = append(res, map[string]interface{}{
			"performance_id": perfID, "student_id": studentID, "grade": grade,
			"status": status, "comment": comment, "has_debt": hasDebt == 1,
			"student_fio": sFio, "group_name": gName, "course": course,
			"discipline_id": discID, "discipline_name": dName, "semester": sem,
			"teacher_id": teacherID, "teacher_fio": tFio, "has_commission": commExists,
		})
	}
	if res == nil { res = emptyArr() }
	sendJSON(w, res)
}

// ==================== COMMISSIONS ====================

func getCommissionsHandler(w http.ResponseWriter, r *http.Request) {
	q := `SELECT c.id, c.performance_id, c.assigned_date, c.meeting_date, c.status, c.decision, c.description,
		s.fio, g.name, d.name, p.grade
		FROM commissions c
		JOIN performance p ON c.performance_id=p.id
		JOIN students s ON p.student_id=s.id JOIN groups_t g ON s.group_id=g.id
		JOIN sheets sh ON p.sheet_id=sh.id JOIN disciplines d ON sh.discipline_id=d.id`

	var conds []string; var args []interface{}
	if v := r.URL.Query().Get("status"); v != "" { conds = append(conds, "c.status=?"); args = append(args, v) }
	if len(conds) > 0 { q += " WHERE " + strings.Join(conds, " AND ") }
	q += " ORDER BY c.assigned_date DESC"

	rows, err := db.Query(q, args...)
	if err != nil { http.Error(w, err.Error(), 500); return }
	defer rows.Close()

	var res []map[string]interface{}
	for rows.Next() {
		var id, perfID int
		var aDate, mDate, status, decision, desc, sFio, gName, dName, grade string
		rows.Scan(&id, &perfID, &aDate, &mDate, &status, &decision, &desc, &sFio, &gName, &dName, &grade)

		// Get members
		mRows, _ := db.Query(`SELECT t.id, t.fio, cm.role FROM commission_members cm
			JOIN teachers t ON cm.teacher_id=t.id WHERE cm.commission_id=? ORDER BY cm.role DESC`, id)
		var members []map[string]interface{}
		for mRows.Next() {
			var tid int; var tFio, role string; mRows.Scan(&tid, &tFio, &role)
			members = append(members, map[string]interface{}{"teacher_id": tid, "fio": tFio, "role": role})
		}
		mRows.Close()
		if members == nil { members = make([]map[string]interface{}, 0) }

		res = append(res, map[string]interface{}{
			"id": id, "performance_id": perfID, "assigned_date": aDate, "meeting_date": mDate,
			"status": status, "decision": decision, "description": desc,
			"student_fio": sFio, "student_group": gName, "discipline_name": dName,
			"grade": grade, "members": members,
		})
	}
	if res == nil { res = emptyArr() }
	sendJSON(w, res)
}

func postCommissionHandler(w http.ResponseWriter, r *http.Request) {
	var c map[string]interface{}; parseJSON(r, &c)
	now := time.Now().Format("2006-01-02")
	aDate := now; if d, ok := c["assigned_date"].(string); ok && d != "" { aDate = d }
	mDate := ""; if d, ok := c["meeting_date"].(string); ok { mDate = d }
	desc := ""; if d, ok := c["description"].(string); ok { desc = d }

	res, err := db.Exec(`INSERT INTO commissions(performance_id,assigned_date,meeting_date,status,decision,description) VALUES(?,?,?,?,?,?)`,
		c["performance_id"], aDate, mDate, "Назначена", "", desc)
	if err != nil { http.Error(w, err.Error(), 500); return }
	commID, _ := res.LastInsertId()

	// Insert members
	if membersI, ok := c["members"]; ok && membersI != nil {
		members := membersI.([]interface{})
		for _, mi := range members {
			m := mi.(map[string]interface{})
			tid := m["teacher_id"]; role := "Член комиссии"
			if r, ok := m["role"].(string); ok && r != "" { role = r }
			db.Exec("INSERT INTO commission_members(commission_id,teacher_id,role) VALUES(?,?,?)", commID, tid, role)
		}
	}
	sendJSON(w, map[string]interface{}{"id": commID})
}

func putCommissionHandler(w http.ResponseWriter, r *http.Request) {
	cid := idFromPath(r); var c map[string]interface{}; parseJSON(r, &c)
	if v, ok := c["status"]; ok { db.Exec("UPDATE commissions SET status=? WHERE id=?", v, cid) }
	if v, ok := c["decision"]; ok { db.Exec("UPDATE commissions SET decision=? WHERE id=?", v, cid) }
	if v, ok := c["description"]; ok { db.Exec("UPDATE commissions SET description=? WHERE id=?", v, cid) }
	if v, ok := c["meeting_date"]; ok { db.Exec("UPDATE commissions SET meeting_date=? WHERE id=?", v, cid) }

	// Replace members if provided
	if membersI, ok := c["members"]; ok && membersI != nil {
		db.Exec("DELETE FROM commission_members WHERE commission_id=?", cid)
		members := membersI.([]interface{})
		for _, mi := range members {
			m := mi.(map[string]interface{})
			tid := m["teacher_id"]; role := "Член комиссии"
			if r, ok := m["role"].(string); ok && r != "" { role = r }
			db.Exec("INSERT INTO commission_members(commission_id,teacher_id,role) VALUES(?,?,?)", cid, tid, role)
		}
	}
	sendJSON(w, map[string]bool{"ok": true})
}

// ==================== HELPERS ====================

func getSemestersHandler(w http.ResponseWriter, r *http.Request) {
	rows, _ := db.Query("SELECT DISTINCT semester FROM disciplines ORDER BY semester")
	defer rows.Close()
	var res []int
	for rows.Next() { var s int; rows.Scan(&s); res = append(res, s) }
	if res == nil { res = make([]int, 0) }
	sendJSON(w, res)
}

func getStudentsByGroupHandler(w http.ResponseWriter, r *http.Request) {
	gid := r.URL.Query().Get("group_id")
	if gid == "" { sendJSON(w, emptyArr()); return }
	rows, err := db.Query("SELECT id, fio, record_book_number FROM students WHERE group_id=? AND status='Активен' ORDER BY fio", gid)
	if err != nil { http.Error(w, err.Error(), 500); return }
	defer rows.Close()
	var res []map[string]interface{}
	for rows.Next() {
		var id int; var fio, rb string; rows.Scan(&id, &fio, &rb)
		res = append(res, map[string]interface{}{"id": id, "fio": fio, "record_book_number": rb})
	}
	if res == nil { res = emptyArr() }
	sendJSON(w, res)
}

func getStatsHandler(w http.ResponseWriter, r *http.Request) {
	var ts, tt, td, tDebt, tComm int
	db.QueryRow("SELECT COUNT(*) FROM students WHERE status='Активен'").Scan(&ts)
	db.QueryRow("SELECT COUNT(*) FROM teachers").Scan(&tt)
	db.QueryRow("SELECT COUNT(*) FROM disciplines").Scan(&td)
	db.QueryRow("SELECT COUNT(*) FROM performance WHERE has_debt=1").Scan(&tDebt)
	db.QueryRow("SELECT COUNT(*) FROM commissions WHERE status!='Закрыта'").Scan(&tComm)
	var avg float64
	db.QueryRow(`SELECT COALESCE(AVG(CASE WHEN grade='Отлично' THEN 5 WHEN grade='Хорошо' THEN 4
		WHEN grade='Удовлетворительно' THEN 3 WHEN grade='Неудовлетворительно' THEN 2
		WHEN grade='Зачтено' THEN 5 WHEN grade='Не зачтено' THEN 2 ELSE 0 END),0) FROM performance`).Scan(&avg)
	sendJSON(w, map[string]interface{}{
		"total_students": ts, "total_teachers": tt, "total_disciplines": td,
		"total_debts": tDebt, "total_commissions": tComm,
		"avg_grade": strconv.FormatFloat(avg, 'f', 2, 64),
	})
}

// ==================== ROUTER ====================

func route(get, post, put, del http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet: if get != nil { get(w, r) } else { sendJSON(w, emptyArr()) }
		case http.MethodPost: if post != nil { post(w, r) } else { w.WriteHeader(405) }
		case http.MethodPut: if put != nil { put(w, r) } else { w.WriteHeader(405) }
		case http.MethodDelete: if del != nil { del(w, r) } else { w.WriteHeader(405) }
		default: w.WriteHeader(405)
		}
	}
}

func main() {
	if err := initDB(filepath.Join(".", "database.sqlite")); err != nil {
		log.Fatalf("DB init failed: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("/api/login", loginHandler)

	// Groups
	mux.HandleFunc("/api/groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet { getGroupsHandler(w, r) }
	})

	// Students
	mux.HandleFunc("/api/students", route(getStudentsHandler, postStudentHandler, nil, nil))
	mux.HandleFunc("/api/students/", route(nil, nil, putStudentHandler, deleteStudentHandler))

	// Teachers
	mux.HandleFunc("/api/teachers", route(getTeachersHandler, postTeacherHandler, nil, nil))
	mux.HandleFunc("/api/teachers/", route(nil, nil, putTeacherHandler, deleteTeacherHandler))

	// Disciplines
	mux.HandleFunc("/api/disciplines", route(getDisciplinesHandler, postDisciplineHandler, nil, nil))
	mux.HandleFunc("/api/disciplines/", route(nil, nil, putDisciplineHandler, deleteDisciplineHandler))

	// Sheets
	mux.HandleFunc("/api/sheets", route(getSheetsHandler, postSheetHandler, nil, nil))

	// Performance
	mux.HandleFunc("/api/performance", route(getPerformanceHandler, postPerformanceHandler, nil, nil))
	mux.HandleFunc("/api/performance/", route(nil, nil, putPerformanceHandler, deletePerformanceHandler))
	mux.HandleFunc("/api/performance/batch", batchPerformanceHandler)

	// Monitoring
	mux.HandleFunc("/api/monitoring", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet { getMonitoringHandler(w, r) }
	})

	// Commissions
	mux.HandleFunc("/api/commissions", route(getCommissionsHandler, postCommissionHandler, nil, nil))
	mux.HandleFunc("/api/commissions/", route(nil, nil, putCommissionHandler, nil))

	// Helpers
	mux.HandleFunc("/api/semesters", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet { getSemestersHandler(w, r) }
	})
	mux.HandleFunc("/api/students-by-group", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet { getStudentsByGroupHandler(w, r) }
	})
	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet { getStatsHandler(w, r) }
	})

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"}, AllowCredentials: false,
	})
	srv := &http.Server{Addr: ":3000", Handler: c.Handler(mux)}
	log.Println("API Server running on http://localhost:3000")
	srv.ListenAndServe()
}
