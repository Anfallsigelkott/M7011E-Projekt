package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"m7011e-projekt/database"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UnsignedResponse struct {
	Message interface{} `json:"message"`
}

func LoginUser(c *gin.Context, db database.Forum_db) {

	bodyAsByteArray, _ := io.ReadAll(c.Request.Body)
	body := make(map[string]string)
	json.Unmarshal(bodyAsByteArray, &body)

	user, err := db.GetUserByUsername(body["username"])
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
	}
	hashedPWbytes := []byte(user.Password)
	plainPWbytes := []byte(body["password"])
	err = bcrypt.CompareHashAndPassword(hashedPWbytes, plainPWbytes)
	if err == nil {
		// create token
		JWT, err := GenerateJWT(user, db)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusOK, JWT)
	} else {
		//return error
		c.IndentedJSON(http.StatusForbidden, "Username and Password do not match")
	}
}

func RegisterUser(c *gin.Context, db database.Forum_db) {
	bodyAsByteArray, _ := io.ReadAll(c.Request.Body)
	body := make(map[string]string)
	json.Unmarshal(bodyAsByteArray, &body)

	// This function should be callable by anyone, no middleware or auth needed

	if !(len(body["password"]) > 0) || !(len(body["username"]) > 0) {
		c.IndentedJSON(http.StatusBadRequest, "Invalid set of arguments was provided")
		return
	}
	tohash := []byte(body["password"])
	hashedPW, err := bcrypt.GenerateFromPassword(tohash, 14)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "Password hashing failed")
		return
	}
	var adminstatus bool
	if len(body["isadmin"]) > 0 {
		adminstatus, err = strconv.ParseBool(body["isadmin"])
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		if adminstatus {
			tokenstring, err := c.Cookie("authtoken")
			user, err := ExtractJWT(tokenstring)
			requester, err := db.GetUserByUsername(user)
			if err != nil {
				c.IndentedJSON(http.StatusInternalServerError, err)
				return
			}
			if !requester.IsAdmin {
				c.IndentedJSON(http.StatusForbidden, gin.H{"error": "Only admins may attempt to create Admin accounts"})
				return
			}
		}
	} else {
		adminstatus = false
	}

	err = db.CreateNewUser(body["username"], string(hashedPW), adminstatus)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err)
	}
	c.IndentedJSON(http.StatusOK, nil)
}

func UpdateUsername(c *gin.Context, db database.Forum_db) {
	bodyAsByteArray, _ := io.ReadAll(c.Request.Body)
	body := make(map[string]string)
	json.Unmarshal(bodyAsByteArray, &body)

	if !(len(body["oldUsername"]) > 0) || !(len(body["newUsername"]) > 0) {
		c.IndentedJSON(http.StatusBadRequest, "Invalid set of arguments was provided")
		return
	}

	tokenstring, err := c.Cookie("authtoken")
	username, err := ExtractJWT(tokenstring)
	userEntry, err := db.GetUserByUsername(username)
	if username != body["oldUsername"] && !userEntry.IsAdmin {
		c.IndentedJSON(http.StatusUnauthorized, gin.H{"error": "Non-admins may not edit others usernames"})
		return
	}

	err = db.UpdateUsername(body["oldUsername"], body["newUsername"])
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err)
	}
	if username != body["oldUsername"] {
		c.IndentedJSON(http.StatusOK, nil)
		return
	} else {
		jwt, _ := GenerateJWT(userEntry, db)
		c.IndentedJSON(http.StatusOK, jwt)
	}
}

// -------------- JWT UTILS -------------- //

func GenerateJWT(user database.Users, db database.Forum_db) (string, error) {
	var mySigningKey = []byte("tempfix")
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["username"] = user.UserName
	//claims["userID"] = user.UserID

	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		fmt_err := fmt.Errorf("Something Went Wrong: %s", err.Error())
		return "", fmt_err
	}
	return tokenString, nil
}

func ValidateJWT(c *gin.Context) { // isValid, fullName, roleID
	//tokenString := c.Param("authtoken")
	tokenString, err := c.Cookie("authtoken")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, UnsignedResponse{
			Message: "no jwt token could be found",
		})
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, OK := token.Method.(*jwt.SigningMethodHMAC); !OK {
			return nil, fmt.Errorf("bad signed method received")
		}
		return []byte("tempfix"), nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusBadRequest, UnsignedResponse{
			Message: "bad jwt token",
		})
		return
	}

	_, OK := token.Claims.(jwt.MapClaims)
	if !OK {
		c.AbortWithStatusJSON(http.StatusInternalServerError, UnsignedResponse{
			Message: "unable to parse claims",
		})
		return
	}
	c.Next()
}

func AdminValidateJWT(c *gin.Context, db database.Forum_db) { // isValid, fullName, roleID
	//tokenString := c.Param("authtoken")
	tokenString, err := c.Cookie("authtoken")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, UnsignedResponse{
			Message: "no jwt token could be found",
		})
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, OK := token.Method.(*jwt.SigningMethodHMAC); !OK {
			return nil, fmt.Errorf("bad signed method received")
		}
		return []byte("tempfix"), nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusBadRequest, UnsignedResponse{
			Message: "bad jwt token",
		})
		return
	}

	_, OK := token.Claims.(jwt.MapClaims)
	if !OK {
		c.AbortWithStatusJSON(http.StatusInternalServerError, UnsignedResponse{
			Message: "unable to parse claims",
		})
		return
	}
	username, _ := ExtractJWT(tokenString)
	user, _ := db.GetUserByUsername(username)
	if !user.IsAdmin {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "unauthorized user"})
		return
	}
	c.Next()
}

func ExtractJWT(tokenString string) (string, error) {
	/*Extraction shouldn't require any further validation or error checking as
	  we perform that in the ValidateJWT function, any errors should be handled there*/
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return os.Getenv("secretkey"), nil
	})

	claims, _ := token.Claims.(jwt.MapClaims)
	return claims["username"].(string), err
}

func AddOriginAdmin(db database.Forum_db) {
	isEmpty, err := db.UserTableIsEmpty()
	if err != nil {
		log.Printf("Internal error when verifying table status. Error: %s\n", err.Error())
		return
	}

	if isEmpty {
		hashedPW, err := bcrypt.GenerateFromPassword([]byte("newpass"), 14)
		if err != nil {
			log.Printf("error generating hashed password: %s\n", err.Error())
			return
		}
		err = db.CreateNewUser("Admin", string(hashedPW), true)
		if err != nil {
			log.Printf("error creating user: %s\n", err.Error())
			return
		}
		log.Println("Origin admin account was created, please use it to create a new account and immediately delete this one")
		return
	}
	log.Println("Table not empty, no account was created")
}
