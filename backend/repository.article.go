// repository.article.go

package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	pb "github.com/fk652/import/commonpb"
)

func (s *server) GetAllArticles(ctx context.Context, args *pb.Request) (*pb.ArticleListReply, error) {
	fmt.Println("GetAllArticles")

	var pbArticles []*pb.Article

	for _, userAList := range articleList {

		userAList.RLock()
		for _, a := range userAList.articleList {
			pbArticles = append(pbArticles, &pb.Article{
				ID:       a.ID,
				User:     a.User,
				PostDate: a.PostDate,
				Title:    a.Title,
				Content:  a.Content,
			})
		}
		userAList.RUnlock()
	}

	sort.Sort(ByTimestamp(pbArticles))
	return &pb.ArticleListReply{Articles: pbArticles}, nil
}

func (s *server) GetSomeArticles(ctx context.Context, args *pb.UsernameRequest) (*pb.ArticleListReply, error) {
	fmt.Println("GetSomeArticles")

	followedUsers, _ := getFollowedUsers2(args.GetUsername())
	var replyList []*pb.Article

	for _, f := range followedUsers {

		userAList := articleList[f]
		userAList.RLock()
		for _, a := range userAList.articleList {
			replyList = append(replyList, &pb.Article{
				ID:       a.ID,
				User:     a.User,
				PostDate: a.PostDate,
				Title:    a.Title,
				Content:  a.Content,
			})
		}
		userAList.RUnlock()
	}

	sort.Sort(ByTimestamp(replyList))
	return &pb.ArticleListReply{Articles: replyList}, nil
}

func (s *server) GetArticleByID(ctx context.Context, args *pb.ArticleIDRequest) (*pb.ArticleReply, error) {
	fmt.Println("GetArticleByID")

	id := args.GetId()

	split := strings.Split(id, "_")
	if len(split) != 2 {
		return nil, errors.New("Invalid Article ID")
	}

	userID, err1 := strconv.Atoi(split[0])
	articleID, err2 := strconv.Atoi(split[1])
	if err1 != nil || err2 != nil || userID < 1 || articleID < 1 {
		return nil, errors.New("Invalid Article ID")
	}

	// validating user ID
	userListRWMutex.RLock()
	if userID >= userIDcount {
		userListRWMutex.RUnlock()
		return nil, errors.New("Invalid Article ID")
	}
	username := userList[userID-1].Username
	userListRWMutex.RUnlock()

	// searching by articleID in the user's article list
	userAList := articleList[username]
	userAList.RLock()
	if articleID >= userAList.articleID {
		userAList.RUnlock()
		return nil, errors.New("Tweet not found")
	}
	a := userAList.articleList[articleID-1]
	userAList.RUnlock()

	pbA := &pb.Article{
		ID:       a.ID,
		User:     a.User,
		PostDate: a.PostDate,
		Title:    a.Title,
		Content:  a.Content,
	}
	return &pb.ArticleReply{Article: pbA}, nil
}

func (s *server) GetArticleByUser(ctx context.Context, args *pb.UsernameRequest) (*pb.ArticleListReply, error) {
	fmt.Println("GetArticleByUser")

	var pbArticles []*pb.Article

	user := args.GetUsername()
	userAList := articleList[user]

	userAList.RLock()

	if len(userAList.articleList) == 0 {
		userAList.RUnlock()
		return nil, errors.New("No tweets found")
	}

	for _, a := range userAList.articleList {
		pbArticles = append(pbArticles, &pb.Article{
			ID:       a.ID,
			User:     a.User,
			PostDate: a.PostDate,
			Title:    a.Title,
			Content:  a.Content,
		})
	}
	userAList.RUnlock()

	sort.Sort(ByTimestamp(pbArticles))
	return &pb.ArticleListReply{Articles: pbArticles}, nil
}

func (s *server) CreateNewArticle(ctx context.Context, args *pb.NewArticleRequest) (*pb.ArticleReply, error) {
	fmt.Println("CreateNewArticle")

	title := args.GetTitle()
	content := args.GetContent()
	user := args.GetUser()
	timestampSeconds := args.GetTimestampSeconds()

	userAList := articleList[user]

	userAList.Lock()
	articleID := strconv.Itoa(userAList.userID) + "_" + strconv.Itoa(userAList.articleID)

	a := article{
		ID:       articleID,
		User:     user,
		PostDate: timestampSeconds,
		Title:    title,
		Content:  content,
	}

	pbA := &pb.Article{
		ID:       articleID,
		User:     user,
		PostDate: timestampSeconds,
		Title:    title,
		Content:  content,
	}

	userAList.articleID++
	userAList.articleList = append(userAList.articleList, a)
	userAList.Unlock()

	return &pb.ArticleReply{Article: pbA}, nil
}

// sorting by timestamp interface
type ByTimestamp []*pb.Article

func (t ByTimestamp) Len() int           { return len(t) }
func (t ByTimestamp) Less(i, j int) bool { return t[i].PostDate > t[j].PostDate }
func (t ByTimestamp) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

// prepend function
func prependArticle(x []*pb.Article, y *pb.Article) []*pb.Article {
	fmt.Println("prependArticle")

	x = append(x, &pb.Article{})
	copy(x[1:], x)
	x[0] = y
	return x
}
