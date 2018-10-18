package main

import (
	"fmt"
	"log"

	"github.com/go-macaron/session"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/macaron.v1"
)

func homeView(ctx *macaron.Context) {
	ctx.HTML(200, "home")
}

func settingsView(ctx *macaron.Context) {
	ctx.HTML(200, "settings")
}

func exchangeView(ctx *macaron.Context) {
	prices, err := pc.DoRequest()
	if err != nil {
		ctx.Data["PriceWav"] = fmt.Sprintf("%.8f", float64(0))
		ctx.Data["PriceBtc"] = fmt.Sprintf("%.8f", float64(0))
		ctx.Data["PriceEth"] = fmt.Sprintf("%.8f", float64(0))
	} else {
		ctx.Data["PriceWav"] = fmt.Sprintf("%.8f", prices.WAVES/100)
		ctx.Data["PriceBtc"] = fmt.Sprintf("%.8f", prices.BTC/100)
		ctx.Data["PriceEth"] = fmt.Sprintf("%.8f", prices.ETH/100)
	}
	ctx.HTML(200, "exchange")
}

func profitView(ctx *macaron.Context) {
	ctx.HTML(200, "profit")
}

func signInView(ctx *macaron.Context) {
	ctx.HTMLSet(200, "login", "signin")
}

func signInPostView(ctx *macaron.Context, siForm SignInForm, sess session.Store) {
	success := &JsonResponse{Success: false}

	user := ctx.Data["User"].(*User)

	if len(user.PasswordHash) == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(siForm.Password), 8)
		if err == nil {
			user.PasswordHash = string(hashedPassword)
			db.Save(user)
		}
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(siForm.Password))
	if err == nil {
		success.Success = true
		sess.Set("userID", user.ID)
	}

	ctx.JSON(200, success)
}

func signOutView(ctx *macaron.Context, sess session.Store) {
	sess.Delete("userID")
	ctx.Redirect("/")
}

func signUpView(ctx *macaron.Context) {
	ctx.HTMLSet(200, "login", "signup")
}

func signUpNewView(ctx *macaron.Context) {
	ctx.HTMLSet(200, "login", "signupnew")
}

func signUpImportView(ctx *macaron.Context) {
	ctx.HTMLSet(200, "login", "signupimport")
}

func signUpPostView(ctx *macaron.Context, siForm SignInForm, sess session.Store) {
	success := &JsonResponse{Success: true}

	user := ctx.Data["User"].(*User)
	log.Println(user)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(siForm.Password), 8)
	if err == nil {
		user.PasswordHash = string(hashedPassword)
		db.Save(user)
	} else {
		success.Success = false
	}

	ctx.JSON(200, success)
}

func localesjsView(ctx *macaron.Context) {
	loc = initLocale(ctx)
	ctx.JSON(200, &loc)
}

func applyView(ctx *macaron.Context, af ApplyForm) {
	success := &JsonResponse{Success: true}
	status := 200
	user := ctx.Data["User"].(*User)

	user.Nickname = af.Nickname
	user.Email = af.Email
	user.Country = af.Country
	user.City = af.City

	err := db.Save(user)

	if err.RowsAffected == 0 {
		success.Success = false
		success.Message = err.Error.Error()
		status = 400
	} else {
		err := sendWelcomeEmail(user)
		if err != nil {
			log.Printf("error sending email: %s", err)
		}
	}

	ctx.JSON(status, success)
}

func verifyView(ctx *macaron.Context, f *session.Flash, sess session.Store) {
	uid, err := decrypt([]byte(conf.DbPass[:16]), ctx.Params("uid"))
	if err != nil {
		return
	}

	log.Println(uid)
	u := &User{Address: uid}
	db.First(u, u)
	u.EmailVerified = true
	db.Save(u)

	var balance uint64
	abr, err := wnc.AssetsBalance(u.Address, "4zbprK67hsa732oSGLB6HzE8Yfdj3BcTcehCeTA1G5Lf")
	if err != nil {
		balance = 0
	} else {
		balance = uint64(abr.Balance)
	}

	log.Printf("balance: %d", balance)

	f.Success("You have successfully verified your email address.")

	if balance >= (1 * satInBtc) {
		applicant := &Badge{Name: "applicant"}
		db.First(applicant)
		db.Model(u).Association("Badges").Append(applicant)

		f.Success("You have successfully applied for Anonutopia citizenship.")
	}

	if balance >= (1000 * satInBtc) {
		citizen := &Badge{Name: "citizen"}
		db.First(citizen)
		db.Model(u).Association("Badges").Append(citizen)
	}

	if balance >= (10000 * satInBtc) {
		founder := &Badge{Name: "funder"}
		db.First(founder)
		db.Model(u).Association("Badges").Append(founder)
	}

	if balance >= (100000 * satInBtc) {
		pioneer := &Badge{Name: "pioneer"}
		db.First(pioneer)
		db.Model(u).Association("Badges").Append(pioneer)
	}

	ctx.Redirect("/settings/")
}
