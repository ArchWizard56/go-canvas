package canvas

import (
	"fmt"
	"path"
	"time"
)

// Course represents a canvas course.
type Course struct {
	ID                   int         `json:"id"`
	Name                 string      `json:"name"`
	SisCourseID          int         `json:"sis_course_id"`
	UUID                 string      `json:"uuid"`
	IntegrationID        interface{} `json:"integration_id"`
	SisImportID          int         `json:"sis_import_id"`
	CourseCode           string      `json:"course_code"`
	WorkflowState        string      `json:"workflow_state"`
	AccountID            int         `json:"account_id"`
	RootAccountID        int         `json:"root_account_id"`
	EnrollmentTermID     int         `json:"enrollment_term_id"`
	GradingStandardID    int         `json:"grading_standard_id"`
	GradePassbackSetting string      `json:"grade_passback_setting"`
	CreatedAt            time.Time   `json:"created_at"`
	StartAt              time.Time   `json:"start_at"`
	EndAt                time.Time   `json:"end_at"`
	Locale               string      `json:"locale"`
	Enrollments          []struct {
		EnrollmentState                string `json:"enrollment_state"`
		Role                           string `json:"role"`
		RoleID                         int64  `json:"role_id"`
		Type                           string `json:"type"`
		UserID                         int64  `json:"user_id"`
		LimitPrivilegesToCourseSection bool   `json:"limit_privileges_to_course_section"`
	} `json:"enrollments"`
	TotalStudents     int         `json:"total_students"`
	Calendar          interface{} `json:"calendar"`
	DefaultView       string      `json:"default_view"`
	SyllabusBody      string      `json:"syllabus_body"`
	NeedsGradingCount int         `json:"needs_grading_count"`

	Term           Term           `json:"term"`
	CourseProgress CourseProgress `json:"course_progress"`

	ApplyAssignmentGroupWeights bool `json:"apply_assignment_group_weights"`
	Permissions                 struct {
		CreateDiscussionTopic bool `json:"create_discussion_topic"`
		CreateAnnouncement    bool `json:"create_announcement"`
	} `json:"permissions"`
	IsPublic                         bool   `json:"is_public"`
	IsPublicToAuthUsers              bool   `json:"is_public_to_auth_users"`
	PublicSyllabus                   bool   `json:"public_syllabus"`
	PublicSyllabusToAuth             bool   `json:"public_syllabus_to_auth"`
	PublicDescription                string `json:"public_description"`
	StorageQuotaMb                   int    `json:"storage_quota_mb"`
	StorageQuotaUsedMb               int    `json:"storage_quota_used_mb"`
	HideFinalGrades                  bool   `json:"hide_final_grades"`
	License                          string `json:"license"`
	AllowStudentAssignmentEdits      bool   `json:"allow_student_assignment_edits"`
	AllowWikiComments                bool   `json:"allow_wiki_comments"`
	AllowStudentForumAttachments     bool   `json:"allow_student_forum_attachments"`
	OpenEnrollment                   bool   `json:"open_enrollment"`
	SelfEnrollment                   bool   `json:"self_enrollment"`
	RestrictEnrollmentsToCourseDates bool   `json:"restrict_enrollments_to_course_dates"`
	CourseFormat                     string `json:"course_format"`
	AccessRestrictedByDate           bool   `json:"access_restricted_by_date"`
	TimeZone                         string `json:"time_zone"`
	Blueprint                        bool   `json:"blueprint"`
	BlueprintRestrictions            struct {
		Content           bool `json:"content"`
		Points            bool `json:"points"`
		DueDates          bool `json:"due_dates"`
		AvailabilityDates bool `json:"availability_dates"`
	} `json:"blueprint_restrictions"`
	BlueprintRestrictionsByObjectType struct {
		Assignment struct {
			Content bool `json:"content"`
			Points  bool `json:"points"`
		} `json:"assignment"`
		WikiPage struct {
			Content bool `json:"content"`
		} `json:"wiki_page"`
	} `json:"blueprint_restrictions_by_object_type"`

	client       *client
	errorHandler func(error, chan int)
}

// Files returns a channel of all the course's files
func (c *Course) Files(opts ...Param) <-chan *File {
	pages := c.pagination(
		c.filespath(),
		filesInitFunc(c.client),
		opts...,
	)
	return onlyFiles(pages, c.errorHandler)
}

// Folders will retrieve the course's folders.
func (c *Course) Folders() <-chan *Folder {
	pages := c.pagination(
		c.folderspath(),
		foldersInitFunc(c.client),
	)
	return onlyFolders(pages, c.errorHandler)
}

// FilesChan will return a channel that sends File structs
// and a channel that sends errors.
func (c *Course) FilesChan() (<-chan *File, <-chan error) {
	p := c.pagination(
		c.filespath(),
		filesInitFunc(c.client),
	)
	_, files, errs := files(p)
	return files, errs
}

// ListFiles returns a slice of files for the course.
func (c *Course) ListFiles(opts ...Param) ([]*File, error) {
	p := c.pagination(
		c.filespath(),
		filesInitFunc(c.client),
		opts...,
	)
	objects, err := p.collect()
	if err != nil {
		return nil, err
	}
	files := make([]*File, len(objects))
	for i, o := range objects {
		files[i] = o.(*File)
	}
	return files, nil
}

// SetErrorHandler will set a error handling callback that is
// used to handle errors in goroutines. The default error handler
// will simply panic.
//
// The callback should accept an error and a quit channel.
// If a value is sent on the quit channel, whatever secsion of
// code is receiving the channel will end gracefully.
func (c *Course) SetErrorHandler(f func(error, chan int)) {
	c.errorHandler = f
}

func (c *Course) pagination(
	path string,
	init pageInitFunction,
	params ...Param,
) *paginated {
	return newPaginatedList(c.client, path, init, params...)
}

// CourseOption is a string type that defines the available course options.
type CourseOption string

const (
	// NeedsGradingCountOpt is a course option
	NeedsGradingCountOpt CourseOption = "needs_grading_count"
	// SyllabusBodyOpt is a course option
	SyllabusBodyOpt CourseOption = "syllabus_body"
	// PublicDescriptionOpt is a course option
	PublicDescriptionOpt CourseOption = "public_description"
	// TotalScoresOpt is a course option
	TotalScoresOpt CourseOption = "total_scores"
	// CurrentGradingPeriodScoresOpt is a course option
	CurrentGradingPeriodScoresOpt CourseOption = "current_grading_period_scores"
	// TermOpt is a course option
	TermOpt CourseOption = "term"
	// AccountOpt is a course option
	AccountOpt CourseOption = "account"
	// CourseProgressOpt is a course option
	CourseProgressOpt CourseOption = "course_progress"
	// SectionsOpt is a course option
	SectionsOpt CourseOption = "sections"
	// StorageQuotaUsedMBOpt is a course option
	StorageQuotaUsedMBOpt CourseOption = "storage_quota_used_mb"
	// TotalStudentsOpt is a course option
	TotalStudentsOpt CourseOption = "total_students"
	// PassbackStatusOpt is a course option
	PassbackStatusOpt CourseOption = "passback_status"
	// FavoritesOpt is a course option
	FavoritesOpt CourseOption = "favorites"
	// TeachersOpt is a course option
	TeachersOpt CourseOption = "teachers"
	// ObservedUsersOpt is a course option
	ObservedUsersOpt CourseOption = "observed_users"
	// CourseImageOpt is a course option
	CourseImageOpt CourseOption = "course_image"
	// ConcludedOpt is a course option
	ConcludedOpt CourseOption = "concluded"
)

func (opt CourseOption) String() string {
	return string(opt)
}

// Term is a school term. One school year.
type Term struct {
	ID      int
	Name    string
	StartAt time.Time `json:"start_at"`
	EndAt   time.Time `json:"end_at"`
}

// CourseProgress is the progress through a course.
type CourseProgress struct {
	RequirementCount          int    `json:"requirement_count"`
	RequirementCompletedCount int    `json:"requirement_completed_count"`
	NextRequirementURL        string `json:"next_requirement_url"`
	CompletedAt               string `json:"completed_at"`
}

// Enrollment is an enrollment object
type Enrollment struct {
	ID                             int         `json:"id"`
	CourseID                       int         `json:"course_id"`
	SisCourseID                    string      `json:"sis_course_id"`
	CourseIntegrationID            string      `json:"course_integration_id"`
	CourseSectionID                int         `json:"course_section_id"`
	SectionIntegrationID           string      `json:"section_integration_id"`
	SisAccountID                   string      `json:"sis_account_id"`
	SisSectionID                   string      `json:"sis_section_id"`
	SisUserID                      string      `json:"sis_user_id"`
	EnrollmentState                string      `json:"enrollment_state"`
	LimitPrivilegesToCourseSection bool        `json:"limit_privileges_to_course_section"`
	SisImportID                    int         `json:"sis_import_id"`
	RootAccountID                  int         `json:"root_account_id"`
	Type                           string      `json:"type"`
	UserID                         int         `json:"user_id"`
	AssociatedUserID               interface{} `json:"associated_user_id"`
	Role                           string      `json:"role"`
	RoleID                         int         `json:"role_id"`
	CreatedAt                      time.Time   `json:"created_at"`
	UpdatedAt                      time.Time   `json:"updated_at"`
	StartAt                        time.Time   `json:"start_at"`
	EndAt                          time.Time   `json:"end_at"`
	LastActivityAt                 time.Time   `json:"last_activity_at"`
	LastAttendedAt                 time.Time   `json:"last_attended_at"`
	TotalActivityTime              int         `json:"total_activity_time"`
	HTMLURL                        string      `json:"html_url"`
	Grades                         struct {
		HTMLURL              string  `json:"html_url"`
		CurrentScore         string  `json:"current_score"`
		CurrentGrade         string  `json:"current_grade"`
		FinalScore           float64 `json:"final_score"`
		FinalGrade           string  `json:"final_grade"`
		UnpostedCurrentGrade string  `json:"unposted_current_grade"`
		UnpostedFinalGrade   string  `json:"unposted_final_grade"`
		UnpostedCurrentScore string  `json:"unposted_current_score"`
		UnpostedFinalScore   string  `json:"unposted_final_score"`
	} `json:"grades"`
	User struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		SortableName string `json:"sortable_name"`
		ShortName    string `json:"short_name"`
	} `json:"user"`
	OverrideGrade                     string  `json:"override_grade"`
	OverrideScore                     float64 `json:"override_score"`
	UnpostedCurrentGrade              string  `json:"unposted_current_grade"`
	UnpostedFinalGrade                string  `json:"unposted_final_grade"`
	UnpostedCurrentScore              string  `json:"unposted_current_score"`
	UnpostedFinalScore                string  `json:"unposted_final_score"`
	HasGradingPeriods                 bool    `json:"has_grading_periods"`
	TotalsForAllGradingPeriodsOption  bool    `json:"totals_for_all_grading_periods_option"`
	CurrentGradingPeriodTitle         string  `json:"current_grading_period_title"`
	CurrentGradingPeriodID            int     `json:"current_grading_period_id"`
	CurrentPeriodOverrideGrade        string  `json:"current_period_override_grade"`
	CurrentPeriodOverrideScore        float64 `json:"current_period_override_score"`
	CurrentPeriodUnpostedCurrentScore float64 `json:"current_period_unposted_current_score"`
	CurrentPeriodUnpostedFinalScore   float64 `json:"current_period_unposted_final_score"`
	CurrentPeriodUnpostedCurrentGrade string  `json:"current_period_unposted_current_grade"`
	CurrentPeriodUnpostedFinalGrade   string  `json:"current_period_unposted_final_grade"`
}

func (c *Course) path(p ...string) string {
	return path.Join(p...)
}

func (c *Course) filespath() string {
	return fmt.Sprintf("courses/%d/files", c.ID)
}

func (c *Course) folderspath() string {
	return fmt.Sprintf("courses/%d/folders", c.ID)
}

func (c *Course) setClient(cl *client) {
	c.client = cl
}

func files(p *paginated) (int, <-chan *File, chan error) {
	files := make(chan *File)
	ch := p.channel()
	go func() {
		for f := range ch {
			files <- f.(*File)
		}
		close(files)
	}()
	return p.n, files, p.errs
}

func folders(p *paginated) (int, <-chan *Folder, chan error) {
	folders := make(chan *Folder)
	ch := p.channel()
	go func() {
		for f := range ch {
			folders <- f.(*Folder)
		}
		close(folders)
	}()
	return p.n, folders, p.errs
}

func onlyFiles(p *paginated, handle func(err error, quit chan int)) <-chan *File {
	results := make(chan *File)
	quit := make(chan int, 1)
	ch := p.channel()
	go func() {
		defer close(results)
		for i := 0; ; i++ {
			select {
			case <-quit:
				return
			case err := <-p.errs:
				if err != nil {
					handle(err, quit)
				}
			case f := <-ch:
				if f == nil {
					return
				}
				results <- f.(*File)
			}
		}
	}()
	return results
}

func onlyFolders(p *paginated, handle func(err error, quit chan int)) <-chan *Folder {
	results := make(chan *Folder)
	quit := make(chan int, 1)
	ch := p.channel()
	go func() {
		defer close(results)
		for i := 0; ; i++ {
			select {
			case <-quit:
				return
			case err := <-p.errs:
				if err != nil {
					handle(err, quit)
				}
			case f := <-ch:
				if f == nil {
					return
				}
				results <- f.(*Folder)
			}
		}
	}()
	return results
}

func defaultErrorHandler(err error, quit chan int) {
	quit <- 1
	close(quit)
	panic(err)
}
