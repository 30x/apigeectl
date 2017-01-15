package utils

// Cluster representation of a target cluster
type Cluster struct {
  Name string
  Cluster string
  SSO string
}

// User representation of a user's credentials
type User struct {
  Username string
  Token string
}

// Context a named combination of user creds and cluster info
type Context struct {
  Name string
  ClusterInfo Cluster
  UserInfo User
  ProxyMgmtApi string
}

// Config shipyardctl configuration object
type Config struct {
  CurrentContext string // name of current Context
  Contexts []Context
}