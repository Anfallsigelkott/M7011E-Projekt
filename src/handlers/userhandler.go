package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"m7011e-projekt/src/database"
	"net/http"
	"os"

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
	err = db.CreateNewUser(body["username"], string(hashedPW))
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err)
	}
	c.IndentedJSON(http.StatusOK, nil)
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

func ExtractJWT(tokenString string) (string, error) {
	/*Extraction shouldn't require any further validation or error checking as
	  we perform that in the ValidateJWT function, any errors should be handled there*/
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return os.Getenv("secretkey"), nil
	})

	claims, _ := token.Claims.(jwt.MapClaims)
	return claims["username"].(string), err
}
