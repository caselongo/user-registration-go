# user-registration-go
A ready-to-use Golang web app template with user registration functionality.

Use this repo as a start for creating your own web app where users can register and login.

I have used this repo myself in this web app: https://www.tour-giro-vuelta.net

To get this code running you only need to:
## 1. implement your user source
Change all functions in the user-source.go file in the root of this repo so that they interact with your user database.
The current code just keeps registered users in memory, but does not save them anywhere externally. 
Sufficient for testing most functionality, but unsuitable for using in a "real" web app.

## 2. enter your smtp credentials 
Fill in your smtp credentials in the user-source.go file in the root of this repo. 
If you do not want users to confirm their e-mail neither to be able to reset their password (not recommended), just set mailSender to nil in main.go.

