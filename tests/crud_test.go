package db

import (
	"context"
	"former/internal/config"
	"former/internal/db"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

var test_db_repo *db.DBRepository
var test_context context.Context
var test_tag = db.Tag{Name: "Test", Description: "Test", Subsystem: "Test", Alias: "Test", IsActive: true}
var test_template = db.Template{Name: "Test", Content: "Test", Subsystem: "Test", IsActive: true, RenderData: map[string]any{}}

func createDBRepository() *db.DBRepository {
	logger := log.New(os.Stdout, "[SERVICE] ", log.LstdFlags)
	app_config := config.LoadConfig()
	db_repo, err := db.SetupDB(context.Background(), app_config.DB.DB_URL, logger)
	if err != nil {
		logger.Fatal(err.Error())
		return nil
	}
	return db_repo
}

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	test_context = context.Background()
	test_db_repo = createDBRepository()
	defer test_db_repo.TearDown()
	exit_code := m.Run()

	os.Exit(exit_code)

}

func TestAddTemplate(t *testing.T) {
	err := test_db_repo.AddTemplate(test_context, &test_template)
	if err != nil || test_template.ID <= 0 {
		t.Errorf("Can't add Template: %s", err.Error())
	}
}

func TestAddTag(t *testing.T) {
	err := test_db_repo.AddTag(
		test_context,
		&test_tag,
	)
	if err != nil || test_tag.ID <= 0 {
		t.Errorf("Can't add Tag: %s", err.Error())
	}
}

func TestGetElementByID(t *testing.T) {
	expected_id := 1 + rand.Intn(4-1)
	dbmodels := []db.DBModel{&db.Template{ID: expected_id}, &db.Tag{ID: expected_id}}
	for _, model := range dbmodels {
		err := test_db_repo.GetElement(
			test_context,
			model,
			db.Columns.ID,
		)
		if err != nil {
			t.Errorf("Error: %s at table %s", err.Error(), model.GetTable())
		}
		if expected_id != model.GetID() {
			t.Errorf(
				"Expected ID: %d got ID: %d at Table: %s",
				expected_id, model.GetID(), model.GetTable(),
			)
		}
	}
}

func TestGetElementByName(t *testing.T) {
	set_data := []struct {
		name  string
		model db.DBModel
	}{
		{"Test", &test_template},
		{"Test", &test_tag},
	}

	for _, data_model := range set_data {
		err := test_db_repo.GetElement(
			test_context,
			data_model.model,
			db.Columns.Name,
		)
		if err != nil {
			t.Errorf("Error: %s at table %s", err.Error(), data_model.model.GetTable())
		}
		if data_model.name != data_model.model.GetName() {
			t.Errorf(
				"Expected Name: %s got Name: %s at Table: %s",
				data_model.name, data_model.model.GetName(), data_model.model.GetTable(),
			)
		}
	}
}
func TestModifyTemplate(t *testing.T) {

	err := test_db_repo.GetElement(test_context, &test_template, db.Columns.Name)
	if err != nil {
		t.Errorf("Error: %s at table %s", err.Error(), test_template.GetTable())
	}

	expected_content := test_template.Content
	test_template.Content = "Test2"
	err = test_db_repo.AddTemplate(test_context, &test_template)

	if err != nil {
		t.Errorf("Can't add Template: %s", err.Error())
	}

	if expected_content == test_template.Content {
		t.Errorf("Can't modify Template: %s", err.Error())
	}
}

func TestModifyTag(t *testing.T) {

	err := test_db_repo.GetElement(test_context, &test_tag, db.Columns.Name)
	if err != nil {
		t.Errorf("Error: %s at table %s", err.Error(), test_tag.GetTable())
	}

	expected_description := test_tag.Description
	test_tag.Description = "Test2"
	err = test_db_repo.AddTag(test_context, &test_tag)

	if err != nil {
		t.Errorf("Can't add Tag: %s", err.Error())
	}

	if expected_description == test_tag.Description {
		t.Errorf("Can't modify Tag: %s", err.Error())
	}
}

func TestGetAllElements(t *testing.T) {
	dbmodels := []db.DBModel{&db.Template{}, &db.Tag{}}
	for _, model := range dbmodels {
		template_arr, err := test_db_repo.GetAllElements(test_context, model, "", true)
		if err != nil {
			t.Errorf("Error: %s at table %s", err.Error(), model.GetTable())
		}
		if len(*template_arr) == 0 {
			t.Error("Empty Output")
		}
	}
}

func TestDeleteElementByName(t *testing.T) {
	set_data := []struct {
		name  string
		model db.DBModel
	}{
		{"Test", &test_template},
		{"Test", &test_tag},
	}
	for _, data_model := range set_data {
		err := test_db_repo.DeleteElement(
			test_context,
			data_model.model,
			db.Columns.Name,
		)
		if err != nil {
			t.Errorf("Error for Delete: %s at table %s", err.Error(), data_model.model.GetTable())
		}

		err = test_db_repo.GetElement(
			test_context,
			data_model.model,
			db.Columns.Name,
		)
		if err == nil {
			t.Errorf("Element wasn't deleted at Table: %s", data_model.model.GetTable())
		}
	}
}

func TestDeleteTagByID(t *testing.T) {
	test_db_repo.AddTag(test_context, &test_tag)

	err := test_db_repo.DeleteElement(
		test_context,
		&test_tag,
		db.Columns.ID,
	)
	if err != nil {
		t.Errorf("Error for Delete Tag: %s", err.Error())
	}
	err = test_db_repo.GetElement(
		test_context,
		&test_tag,
		db.Columns.ID,
	)
	if err == nil {
		t.Error("Tag wasn't deleted")
	}
}

func TestDeleteTemplateByID(t *testing.T) {
	test_db_repo.AddTemplate(test_context, &test_template)
	err := test_db_repo.DeleteElement(
		test_context,
		&test_template,
		db.Columns.ID,
	)
	if err != nil {
		t.Errorf("Error for Delete Template: %s %d", err.Error(), test_template.ID)
	}
	err = test_db_repo.GetElement(
		test_context,
		&test_template,
		db.Columns.ID,
	)
	if err == nil {
		t.Error("Template wasn't deleted")
	}
}
