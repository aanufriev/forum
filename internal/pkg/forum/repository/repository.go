package repository

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/aanufriev/forum/internal/pkg/forum"
	"github.com/aanufriev/forum/internal/pkg/models"
	"github.com/go-openapi/strfmt"
)

type ForumRepository struct {
	db *sql.DB
}

func New(db *sql.DB) forum.Repository {
	return ForumRepository{
		db: db,
	}
}

func (f ForumRepository) Create(model models.Forum) error {
	_, err := f.db.Exec(
		"INSERT INTO forums (slug, title, user_nickname) VALUES($1, $2, $3)",
		model.Slug, model.Title, model.User,
	)

	if err != nil {
		return fmt.Errorf("couldn't create new forum. Error: %w", err)
	}

	return nil
}

func (f ForumRepository) Get(slug string) (models.Forum, error) {
	var model models.Forum
	err := f.db.QueryRow(
		`SELECT slug, title, user_nickname, thread_count, post_count FROM forums
		WHERE slug = $1`,
		slug,
	).Scan(&model.Slug, &model.Title, &model.User, &model.Threads, &model.Posts)

	if err != nil {
		return models.Forum{}, fmt.Errorf("couldn't get forum with slug '%v'. Error: %w", slug, err)
	}

	return model, nil
}

func (f ForumRepository) CreateThread(thread *models.Thread) error {
	var err error
	thread.Forum, err = f.CheckForum(thread.Forum)

	if err != nil {
		return forum.ErrForumDoesntExists
	}

	err = f.db.QueryRow(
		`INSERT INTO threads (author, created, forum, msg, title, slug)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		thread.Author, thread.Created, thread.Forum, thread.Message, thread.Title, thread.Slug,
	).Scan(&thread.ID)

	if err != nil {
		return fmt.Errorf("couldn't create thread. Error: %w", err)
	}

	return nil
}

func (f ForumRepository) CheckForum(slug string) (string, error) {
	err := f.db.QueryRow(
		"SELECT slug FROM forums WHERE slug = $1",
		slug,
	).Scan(&slug)

	if err != nil {
		return "", fmt.Errorf("couldn't get forum with slug '%v'. Error: %w", slug, err)
	}

	return slug, nil
}

func (f ForumRepository) GetThreads(slug string, limit string, since string, desc string) ([]models.Thread, error) {
	query := "SELECT author, created, forum, id, msg, slug, title, votes FROM threads WHERE forum = $1"

	args := make([]interface{}, 0, 4)
	args = append(args, slug)

	var operator string
	if desc == "" || desc == "false" {
		operator = ">"
	} else {
		operator = "<"
	}

	if since != "" {
		query += fmt.Sprintf(" AND created %v= $2", operator)
		args = append(args, since)
	}

	if desc == "" || desc == "false" {
		desc = "ASC"
	} else {
		desc = "DESC"
	}
	query += fmt.Sprintf(" ORDER BY created %v", desc)

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return nil, err
	}
	query += fmt.Sprintf(" LIMIT %v", limitInt)

	rows, err := f.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	threads := make([]models.Thread, 0, limitInt)
	var thread models.Thread
	for rows.Next() {
		err = rows.Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
		if err != nil {
			return nil, err
		}

		threads = append(threads, thread)
	}

	return threads, nil
}

func (f ForumRepository) CreatePosts(slug string, id int, posts []models.Post) ([]models.Post, error) {
	var forum string
	var slugNull sql.NullString

	err := f.db.QueryRow(
		`SELECT forum, id, slug FROM threads
		WHERE slug = $1 OR id = $2`,
		slug, id,
	).Scan(&forum, &id, &slugNull)

	if err != nil {
		return nil, err
	}

	if len(posts) > 0 && posts[0].Parent != 0 {
		parentPost, err := f.GetPostDetails(strconv.Itoa(posts[0].Parent))
		if err != nil {
			return nil, err
		}

		if parentPost.Thread != id {
			return nil, fmt.Errorf("wrong parent")
		}
	}

	created := strfmt.DateTime(time.Now())
	for idx := range posts {
		posts[idx].Forum = forum
		posts[idx].Thread = id
		posts[idx].Slug = slugNull.String
		posts[idx].Created = created

		err := f.db.QueryRow(`
			INSERT INTO posts (author, msg, parent, thread, thread_slug, forum, created)
			VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
			posts[idx].Author, posts[idx].Message, posts[idx].Parent, posts[idx].Thread, posts[idx].Slug, posts[idx].Forum, posts[idx].Created,
		).Scan(&posts[idx].ID)

		if err != nil {
			return nil, err
		}
	}

	return posts, nil
}

func (f ForumRepository) GetThread(slug string, id int) (models.Thread, error) {
	var thread models.Thread
	err := f.db.QueryRow(
		`SELECT author, created, forum, id, msg, slug, title, votes FROM threads
		WHERE slug = $1 or id = $2`,
		slug, id,
	).Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)

	if err != nil {
		return models.Thread{}, err
	}

	return thread, nil
}

func (f ForumRepository) Vote(vote models.Vote) (models.Thread, error) {
	thread, err := f.GetThread(vote.Slug, vote.ID)
	if err != nil {
		return models.Thread{}, err
	}

	var voteValue int
	err = f.db.QueryRow(
		`SELECT vote FROM thread_vote
		WHERE nickname = $1 AND thread_id = $2`,
		vote.Nickname, thread.ID,
	).Scan(&voteValue)

	if err != nil && err != sql.ErrNoRows {
		return models.Thread{}, err
	}

	if err == sql.ErrNoRows {
		_, err = f.db.Exec(
			"INSERT INTO thread_vote (nickname, thread_id, vote) VALUES($1, $2, $3)",
			vote.Nickname, thread.ID, vote.Voice,
		)

		if err != nil {
			return models.Thread{}, err
		}

		thread.Votes += vote.Voice
		return thread, nil
	}

	if voteValue == vote.Voice {
		return thread, nil
	}

	thread.Votes = thread.Votes - voteValue + vote.Voice

	_, err = f.db.Exec(
		`UPDATE thread_vote SET vote = $1
		WHERE nickname = $2 AND thread_id = $3`,
		vote.Voice, vote.Nickname, thread.ID,
	)

	if err != nil {
		return models.Thread{}, err
	}

	return thread, nil
}

func (f ForumRepository) GetPosts(slug string, id int, limit int, order string, since string) ([]models.Post, error) {
	var sinceCond string
	if since != "" {
		if order == "DESC" {
			sinceCond = fmt.Sprintf("AND id < %v", since)
		} else {
			sinceCond = fmt.Sprintf("AND id > %v", since)
		}
	}
	rows, err := f.db.Query(
		fmt.Sprintf(`SELECT author, created, forum, id, msg, parent, thread FROM posts
		WHERE (thread_slug = $1 OR thread = $2) %v
		ORDER BY id %v`, sinceCond, order),
		slug, id,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]models.Post, 0, limit)
	post := models.Post{}
	idx := 0
	for rows.Next() {
		if idx == limit {
			break
		}
		idx++
		err = rows.Scan(&post.Author, &post.Created, &post.Forum, &post.ID, &post.Message, &post.Parent, &post.Thread)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func (f ForumRepository) GetPostsTree(slug string, id int, limit int, order string, since string) ([]models.Post, error) {
	var rows *sql.Rows
	var err error

	var desc bool
	if order == "DESC" {
		desc = true
	} else {
		desc = false
	}

	if since == "" {
		if desc {
			rows, err = f.db.Query(
				`SELECT author, created, forum, id, msg, parent, thread FROM posts
				WHERE (thread_slug = $1 OR thread = $2) ORDER BY path DESC, id  DESC LIMIT $3;`,
				slug, id, limit,
			)
		} else {
			rows, err = f.db.Query(
				`SELECT author, created, forum, id, msg, parent, thread FROM posts
				WHERE (thread_slug = $1 OR thread = $2) ORDER BY path ASC, id  ASC LIMIT $3;`,
				slug, id, limit,
			)
		}
	} else {
		if desc {
			rows, err = f.db.Query(
				`SELECT author, created, forum, id, msg, parent, thread FROM posts
				WHERE (thread_slug = $1 OR thread = $2) AND PATH < (SELECT path FROM posts WHERE id = $3)
				ORDER BY path DESC, id  DESC LIMIT $4;`,
				slug, id, since, limit,
			)
		} else {
			rows, err = f.db.Query(
				`SELECT author, created, forum, id, msg, parent, thread FROM posts
				WHERE (thread_slug = $1 OR thread = $2) AND PATH > (SELECT path FROM posts WHERE id = $3)
				ORDER BY path ASC, id  ASC LIMIT $4;`,
				slug, id, since, limit,
			)
		}
	}
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	posts := make([]models.Post, 0)
	var post models.Post
	for rows.Next() {
		err = rows.Scan(
			&post.Author, &post.Created, &post.Forum, &post.ID,
			&post.Message, &post.Parent, &post.Thread,
		)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func (f ForumRepository) GetPostsParentTree(slug string, id int, limit int, order string, since string) ([]models.Post, error) {
	var rows *sql.Rows
	var err error

	var desc bool
	if order == "DESC" {
		desc = true
	} else {
		desc = false
	}

	if since == "" {
		if desc {
			rows, err = f.db.Query(
				`SELECT author, created, forum, id, msg, parent, thread FROM posts
				WHERE path[1] IN (SELECT id FROM posts WHERE (thread_slug = $1 OR thread = $2) AND parent = 0 ORDER BY id DESC LIMIT $3)
				ORDER BY path[1] DESC, path, id;`,
				slug, id, limit,
			)
		} else {
			rows, err = f.db.Query(
				`SELECT author, created, forum, id, msg, parent, thread FROM posts
				WHERE path[1] IN (SELECT id FROM posts WHERE (thread_slug = $1 OR thread = $2) AND parent = 0 ORDER BY id LIMIT $3)
				ORDER BY path, id;`,
				slug, id, limit,
			)
		}
	} else {
		if desc {
			rows, err = f.db.Query(
				`SELECT author, created, forum, id, msg, parent, thread FROM posts
				WHERE path[1] IN (SELECT id FROM posts WHERE (thread_slug = $1 OR thread = $2) AND parent = 0 AND path[1] <
				(SELECT path[1] FROM posts WHERE id = $3) ORDER BY id DESC LIMIT $4) ORDER BY path[1] DESC, path, id;`,
				slug, id, since, limit,
			)
		} else {
			rows, err = f.db.Query(
				`SELECT author, created, forum, id, msg, parent, thread FROM posts
				WHERE path[1] IN (SELECT id FROM posts WHERE (thread_slug = $1 OR thread = $2) AND parent = 0 AND path[1] >
				(SELECT path[1] FROM posts WHERE id = $3) ORDER BY id ASC LIMIT $4) ORDER BY path, id;`,
				slug, id, since, limit,
			)
		}
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	posts := make([]models.Post, 0)
	var post models.Post
	for rows.Next() {
		err = rows.Scan(
			&post.Author, &post.Created, &post.Forum, &post.ID,
			&post.Message, &post.Parent, &post.Thread,
		)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func (f ForumRepository) UpdateThread(thread models.Thread) (models.Thread, error) {
	if thread.Title == "" || thread.Message == "" {
		oldThread, err := f.GetThread(*thread.Slug, thread.ID)
		if err != nil {
			return models.Thread{}, nil
		}

		if thread.Title == "" {
			thread.Title = oldThread.Title
		}

		if thread.Message == "" {
			thread.Message = oldThread.Message
		}
	}

	err := f.db.QueryRow(
		`UPDATE threads SET title = $1, msg = $2
		WHERE slug = $3 OR id = $4
		RETURNING author, created, forum, id, msg, slug, title`,
		thread.Title, thread.Message, thread.Slug, thread.ID,
	).Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &thread.Slug, &thread.Title)

	if err != nil {
		return models.Thread{}, err
	}

	return thread, nil
}

func (f ForumRepository) GetUsersFromForum(slug string, limit int, since string, desc string) ([]models.User, error) {
	var (
		rows *sql.Rows
		err  error
	)

	var compare string
	if desc == "DESC" {
		compare = "<"
	} else {
		compare = ">"
	}

	if since != "" {
		if limit != 0 {
			rows, err = f.db.Query(
				fmt.Sprintf(`SELECT u.about, u.email, u.fullname, u.nickname FROM users AS u
				JOIN forum_user AS fu ON u.nickname = fu.nickname
				WHERE fu.forum_slug = $1 AND fu.nickname %v $2
				ORDER BY u.nickname %v
				LIMIT %v`, compare, desc, limit),
				slug, since,
			)
		} else {
			rows, err = f.db.Query(
				fmt.Sprintf(`SELECT u.about, u.email, u.fullname, u.nickname FROM users AS u
				JOIN forum_user AS fu ON u.nickname = fu.nickname
				WHERE fu.forum_slug = $1 AND fu.nickname %v $2
				ORDER BY u.nickname %v`, compare, desc),
				slug, since,
			)
		}
	} else {
		if limit != 0 {
			rows, err = f.db.Query(
				fmt.Sprintf(`SELECT u.about, u.email, u.fullname, u.nickname FROM users AS u
				JOIN forum_user AS fu ON u.nickname = fu.nickname
				WHERE fu.forum_slug = $1
				ORDER BY u.nickname %v
				LIMIT %v`, desc, limit),
				slug,
			)
		} else {
			rows, err = f.db.Query(
				fmt.Sprintf(`SELECT u.about, u.email, u.fullname, u.nickname FROM users AS u
				JOIN forum_user AS fu ON u.nickname = fu.nickname
				WHERE fu.forum_slug = $1
				ORDER BY u.nickname %v`, desc),
				slug,
			)
		}
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0, limit)
	user := models.User{}
	for rows.Next() {
		err = rows.Scan(&user.About, &user.Email, &user.Fullname, &user.Nickname)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (f ForumRepository) GetPostDetails(id string) (models.Post, error) {
	var post models.Post
	err := f.db.QueryRow(
		"SELECT author, created, forum, id, msg, thread, isEdited, parent FROM posts WHERE id = $1",
		id,
	).Scan(&post.Author, &post.Created, &post.Forum, &post.ID, &post.Message, &post.Thread, &post.IsEdited, &post.Parent)

	if err != nil {
		return models.Post{}, err
	}

	return post, nil
}

func (f ForumRepository) UpdatePost(post models.Post) (models.Post, error) {
	postDB, err := f.GetPostDetails(strconv.Itoa(post.ID))
	if err != nil {
		return models.Post{}, err
	}

	if post.Message == postDB.Message {
		return postDB, nil
	}

	err = f.db.QueryRow(
		`UPDATE posts SET msg = $1, isEdited = true WHERE id = $2
		RETURNING author, created, forum, id, msg, thread, isEdited, parent`,
		post.Message, post.ID,
	).Scan(&post.Author, &post.Created, &post.Forum, &post.ID, &post.Message, &post.Thread, &post.IsEdited, &post.Parent)

	if err != nil {
		return models.Post{}, err
	}

	return post, nil
}

func (f ForumRepository) ClearService() error {
	_, err := f.db.Exec(
		"TRUNCATE TABLE users, forums, forum_user, threads, thread_vote, posts",
	)

	if err != nil {
		return err
	}

	return nil
}

func (f ForumRepository) GetServiceInfo() (models.ServiceInfo, error) {
	var info models.ServiceInfo
	err := f.db.QueryRow(
		"SELECT count(*) FROM forums",
	).Scan(&info.Forum)

	if err != nil {
		return models.ServiceInfo{}, err
	}

	err = f.db.QueryRow(
		"SELECT count(*) FROM threads",
	).Scan(&info.Thread)

	if err != nil {
		return models.ServiceInfo{}, err
	}

	err = f.db.QueryRow(
		"SELECT count(*) FROM posts",
	).Scan(&info.Post)

	if err != nil {
		return models.ServiceInfo{}, err
	}

	err = f.db.QueryRow(
		"SELECT count(*) FROM users",
	).Scan(&info.User)

	if err != nil {
		return models.ServiceInfo{}, err
	}

	return info, nil
}

func (f ForumRepository) CheckThread(slug string, id int) error {
	err := f.db.QueryRow(
		"SELECT id FROM threads WHERE slug = $1 OR id = $2",
		slug, id,
	).Scan(&id)

	return err
}

func (f ForumRepository) GetThreadIDBySlug(slug string) (int, error) {
	var threadID int
	err := f.db.QueryRow(
		"SELECT id FROM threads WHERE slug = $1",
		slug,
	).Scan(&threadID)

	if err != nil {
		return 0, err
	}

	return threadID, nil
}
