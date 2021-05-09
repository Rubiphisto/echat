package sessions

var (
	userManager = UserManager{users: map[string]*User{}}
)

type UserManager struct {
	users		map[string]*User
}


func GetUserManager() *UserManager {
	return &userManager
}

func (m *UserManager) CreateUser(username string, session *Session) *User {
	_, ok := m.users[username]
	if ok {
		// username is already exist
		return nil
	}
	user := &User{
		userName:  username,
		session:   session,
	}
	m.users[username] = user
	return user
}

func (m *UserManager) GetUser(username string) *User {
	user, ok := m.users[username]
	if !ok {
		return nil
	}
	return user
}

func (m *UserManager) RemoveUser(username string) *User {
	user, ok := m.users[username]
	if !ok {
		return nil
	}
	delete(m.users, username)
	return user
}

