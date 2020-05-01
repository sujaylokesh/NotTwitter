// repository.user.go

package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	pb "github.com/fk652/import/commonpb"
)

func (s *server) IsUserValid(ctx context.Context, args *pb.AccountRequest) (*pb.BoolReply, error) {
	fmt.Println("IsUserValid")

	username := args.GetUsername()
	password := args.GetPassword()

	userListRWMutex.RLock()
	defer userListRWMutex.RUnlock()

	for _, u := range userList {
		if u.Username == username && u.Password == password {
			return &pb.BoolReply{Reply: true}, nil
		}
	}
	return &pb.BoolReply{Reply: false}, nil
}

func (s *server) RegisterNewUser(ctx context.Context, args *pb.AccountRequest) (*pb.BoolReply, error) {
	fmt.Println("RegisterNewUser")

	username := args.GetUsername()
	password := args.GetPassword()

	if strings.TrimSpace(password) == "" {
		return &pb.BoolReply{Reply: false}, errors.New("The password can't be empty")
	} else if !isUsernameAvailable2(username) {
		return &pb.BoolReply{Reply: false}, errors.New("The username isn't available")
	}

	userListRWMutex.Lock()
	u := user{Username: username, Password: password, UserID: userIDcount}
	userList = append(userList, u)
	userIDcount++
	userListRWMutex.Unlock()

	followsRWMutex.Lock()
	follows[username] = &userFollowList{}
	followsRWMutex.Unlock()

	articleRWMutex.Lock()
	articleList[username] = &userArticleList{articleID: 1, userID: u.UserID}
	articleRWMutex.Unlock()

	return &pb.BoolReply{Reply: true}, nil
}

func (s *server) IsUsernameAvailable(ctx context.Context, args *pb.UsernameRequest) (*pb.BoolReply, error) {
	fmt.Println("IsUsernameAvailable")

	username := args.GetUsername()

	userListRWMutex.RLock()
	defer userListRWMutex.RUnlock()

	for _, u := range userList {
		if u.Username == username {
			return &pb.BoolReply{Reply: false}, nil
		}
	}
	return &pb.BoolReply{Reply: true}, nil
}

func (s *server) IsFollowed(ctx context.Context, args *pb.FollowRequest) (*pb.IsFollowedReply, error) {
	fmt.Println("IsFollowed")

	thisUser := args.GetThisUser()
	followUser := args.GetFollowUser()

	if isUsernameAvailable2(followUser) {
		return &pb.IsFollowedReply{Found: false, Index: -1}, errors.New("The user doesn't exist")
	}

	usernames := follows[thisUser]
	usernames.RLock()
	defer usernames.RUnlock()

	for i, a := range usernames.followList {
		if a == followUser {
			return &pb.IsFollowedReply{Found: true, Index: int64(i)}, nil
		}
	}

	return &pb.IsFollowedReply{Found: false, Index: -1}, nil
}

func (s *server) AddFollow(ctx context.Context, args *pb.FollowRequest) (*pb.Reply, error) {
	fmt.Println("AddFollow")

	followUser := args.GetFollowUser()
	thisUser := args.GetThisUser()
	found, _, err := isFollowed2(thisUser, followUser)

	if err != nil {
		return &pb.Reply{Message: "error"}, err
	}

	if found {
		return &pb.Reply{Message: "error"}, errors.New("This user is already followed")
	}

	f := follows[thisUser]

	f.Lock()
	f.followList = append(f.followList, followUser)
	f.Unlock()

	return &pb.Reply{Message: "success"}, nil
}

func (s *server) RemoveFollow(ctx context.Context, args *pb.FollowRequest) (*pb.Reply, error) {
	fmt.Println("RemoveFollow")

	thisUser := args.GetThisUser()
	followUser := args.GetFollowUser()

	found, i, err := isFollowed2(thisUser, followUser)

	if err != nil {
		return &pb.Reply{Message: "error"}, err
	}

	if !found {
		return &pb.Reply{Message: "error"}, errors.New("This user isn't followed")
	}

	f := follows[thisUser]

	f.Lock()
	f.followList = append(f.followList[:i], f.followList[i+1:]...)
	f.Unlock()

	return &pb.Reply{Message: "success"}, nil
}

func (s *server) GetFollowedUsers(ctx context.Context, args *pb.UsernameRequest) (*pb.UsernameListReply, error) {
	fmt.Println("GetFollowedUsers")

	thisUser := args.GetUsername()

	if thisUser == "" {
		return &pb.UsernameListReply{FollowList: []string{}}, errors.New("The user of this account not found")
	}

	f := follows[thisUser]
	f.RLock()
	defer f.RUnlock()

	return &pb.UsernameListReply{FollowList: f.followList}, nil
}

// Non-RPC versions for use within backend functions
func isUsernameAvailable2(username string) bool {
	fmt.Println("isUsernameAvailable2")

	userListRWMutex.RLock()
	defer userListRWMutex.RUnlock()

	for _, u := range userList {
		if u.Username == username {
			return false
		}
	}
	return true
}

func isFollowed2(thisUser string, followUser string) (bool, int, error) {
	fmt.Println("isFollowed2")

	if isUsernameAvailable2(followUser) {
		return false, -1, errors.New("The user doesn't exist")
	}

	usernames := follows[thisUser]
	usernames.RLock()
	defer usernames.RUnlock()

	for i, a := range usernames.followList {
		if a == followUser {
			return true, i, nil
		}
	}

	return false, -1, nil
}

func getFollowedUsers2(username string) ([]string, error) {
	fmt.Println("getFollowedUsers2")

	if username == "" {
		return nil, errors.New("The user of this account not found")
	}

	f := follows[username]
	f.RLock()
	defer f.RUnlock()

	return f.followList, nil
}
