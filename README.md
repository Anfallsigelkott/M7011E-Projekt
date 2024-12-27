# M7011E-Projekt
A project for the course M7011E Design of Dynamic Web Systems. 

The project is intended to be a forum client where users can create accounts, create groups, and then submit posts to these groups. A user can see all groups but can only see the content of groups they have joined. Within groups there are two roles, one for a regular user and then one for a moderator which has authority to delete others posts and manage users within the group. External to the groups there is also a admin role which functions as a moderator for the entire service instead of for just one group and can perform the most sensitive actions like deleting groups.

# Setup
The project has been developed using docker-composer/docker desktop and it is what we recommend to use in order to run it.
To run the project simply start the docker engine on your computer and then use a terminal to write `docker-compose up --build -d` while in the appropriate directory, the API should now be reachable through clients such as postman at localhost:8080.

# API documentation

Prepend each route with /v1 (so use localhost:8080/v1/groups to get the groups for example).

## GET

- /groups
    - Returns a list of all group IDs and names.
- /groups/user/:id 
    - Returns a list of the group IDs and names that a user with the provided ID is part of.
- /groups/:id/user
    - Gives a list of users (usernames) that are part of the group with the provided ID and their role within the group.
- /groups/:id/post
    - Fetches all posts that have been posted in the group with the provided group ID.


## POST

### Creators

- /user/new
    - Body: `{"username": "<username>", "password": "<password>", "isadmin": "<true/false>"}`
    - Registers a new user with the provided username and password. The password is hashed. If isadmin is set to true the created user will be an admin, but that is *only* allowed if the user making the post request has a token that says that they are also an admin.
- /group/new/:group
    - Creates a new group with :group as the new groups name.
- /post/:group
    - Body: `{"content": "<text of post>", "replyID": "<postID this post is in reply to, can be left blank>"}`
    - Adds a new post to the group with the ID :group that contains the content provided in the request body. Can be a reply to another post. Which user posted it is taken from the users authorisation token.
- /user/:user/join/:group/
    - A user with the ID :user is made a part of the group with the ID :group.

### Updaters

- /user/:user/role/:groupID/:newRole
    - Gives the user with the ID :user a new role (:newRole) in the group with ID :groupID. Only a moderator or admin is allowed to promote another user to moderator.
- /post/:group/edit
    - Body: `{"content": "<new text of post>", "postID": "<which post ID that is to be edited>"}`
    - Edits the post with the provided post ID in the group with the ID :group to the new content. Only the user that posted the post or a moderator/admin is allowed to edit a post.
- /user/edit
    - Body `{"oldUsername": "<old username>", "newUsername": "<new username>"}`
    - Updates the username of ones user from the old username to the new username. Only the user themselves and admins are allowed to do this.
- /user/login
    - Body `{"username": "<username>", "password": "<password>"}`
    - Logs in the user if the provided username and password are correct. Returns a JWT cookie that keeps the user logged in and keeps track of their username and admin status.
    - To use the cookie in a tool like postman; press the "Cookies" button and type in localhost as your domain name, then press "Add", after this press "Add cookie" and edit it to look like this `authtoken="<your token>"; <there will be additional data here from postman you don't need to worry about>`. Whenever you get a new cookie simply replace the string within the quotes with your new cookie.

### Removers

- /user/:user/leave/:group
    - Removes a user with the ID :user from the group with the ID :group. Only the user in question, a moderator of the group in question or an admin is allowed to do this.    
- /group/:group/delete
    - Deletes the group with the ID :group. Only accessible to admins.


