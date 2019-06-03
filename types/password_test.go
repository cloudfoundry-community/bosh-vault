package types_test

import (
	"encoding/json"
	"github.com/cloudfoundry-community/bosh-vault/store"
	"github.com/cloudfoundry-community/bosh-vault/types"
	fakes "github.com/cloudfoundry-community/bosh-vault/types/typesfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"unicode"
)

var _ = Describe("Password", func() {
	Describe("request validation", func() {
		Context("a valid password post request", func() {
			var (
				PasswordRequest types.PasswordRequest
			)

			BeforeEach(func() {
				err := json.Unmarshal([]byte(fakes.PasswordPostRequestBody), &PasswordRequest)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns true when Validate is called", func() {
				Expect(PasswordRequest.Validate()).To(BeTrue())
			})
		})
	})
	Describe("generation", func() {
		Context("parameter usage", func() {
			It("respects the length parameter", func() {
				pr := types.PasswordRequest{
					Name: "some_pass",
					Type: types.PasswordType,
					Parameters: types.PasswordParams{
						Length:         100,
						ExcludeUpper:   false,
						ExcludeLower:   false,
						ExcludeNumber:  false,
						IncludeSpecial: true,
					},
				}
				// generate 100 passwords
				for i := 0; i < 100; i++ {
					cr, err := pr.Generate(&store.SimpleStore{})
					Expect(err).ToNot(HaveOccurred())
					Expect(cr.(types.PasswordRecord)).ToNot(BeEmpty())
					Expect(len(cr.(types.PasswordRecord))).To(Equal(100))
				}
			})
			It("generates a password with all the alphabets", func() {
				pr := types.PasswordRequest{
					Name: "some_pass",
					Type: types.PasswordType,
					Parameters: types.PasswordParams{
						Length:         100, // plenty long to include something from all 4 alphabets
						ExcludeUpper:   false,
						ExcludeLower:   false,
						ExcludeNumber:  false,
						IncludeSpecial: true,
					},
				}
				// generate 100 passwords
				for i := 0; i < 100; i++ {
					cr, err := pr.Generate(&store.SimpleStore{})
					Expect(err).ToNot(HaveOccurred())
					Expect(cr.(types.PasswordRecord)).ToNot(BeEmpty())
					var seenLower, seenUpper, seenSpecial, seenNumber bool
					for _, char := range cr.(types.PasswordRecord) {
						if unicode.IsLower(char) {
							seenLower = true
						}
						if unicode.IsUpper(char) {
							seenUpper = true
						}
						if unicode.IsNumber(char) {
							seenNumber = true
						}
						if !unicode.IsNumber(char) && !unicode.IsUpper(char) && !unicode.IsLower(char) {
							seenSpecial = true
						}
					}
					Expect(seenLower && seenUpper && seenNumber && seenSpecial).To(BeTrue())
				}
			})
			It("generates a password with no lowercase letters", func() {
				pr := types.PasswordRequest{
					Name: "some_pass",
					Type: types.PasswordType,
					Parameters: types.PasswordParams{
						Length:         0,
						ExcludeUpper:   false,
						ExcludeLower:   true,
						ExcludeNumber:  false,
						IncludeSpecial: true,
					},
				}
				// generate 100 passwords
				for i := 0; i < 100; i++ {
					cr, err := pr.Generate(&store.SimpleStore{})
					Expect(err).ToNot(HaveOccurred())
					Expect(cr.(types.PasswordRecord)).ToNot(BeEmpty())
					for _, char := range cr.(types.PasswordRecord) {
						Expect(unicode.IsLower(char)).To(BeFalse())
					}
				}
			})
			It("generates a password with only lowercase letters", func() {
				pr := types.PasswordRequest{
					Name: "some_pass",
					Type: types.PasswordType,
					Parameters: types.PasswordParams{
						Length:         0,
						ExcludeUpper:   true,
						ExcludeLower:   false,
						ExcludeNumber:  true,
						IncludeSpecial: false,
					},
				}
				// generate 100 passwords
				for i := 0; i < 100; i++ {
					cr, err := pr.Generate(&store.SimpleStore{})
					Expect(err).ToNot(HaveOccurred())
					Expect(cr.(types.PasswordRecord)).ToNot(BeEmpty())
					for _, char := range cr.(types.PasswordRecord) {
						Expect(unicode.IsLower(char)).To(BeTrue())
					}
				}
			})
			It("generates a password with no uppercase letters", func() {
				pr := types.PasswordRequest{
					Name: "some_pass",
					Type: types.PasswordType,
					Parameters: types.PasswordParams{
						Length:         0,
						ExcludeUpper:   true,
						ExcludeLower:   false,
						ExcludeNumber:  false,
						IncludeSpecial: true,
					},
				}
				// generate 100 passwords
				for i := 0; i < 100; i++ {
					cr, err := pr.Generate(&store.SimpleStore{})
					Expect(err).ToNot(HaveOccurred())
					Expect(cr.(types.PasswordRecord)).ToNot(BeEmpty())
					for _, char := range cr.(types.PasswordRecord) {
						Expect(unicode.IsUpper(char)).To(BeFalse())
					}
				}
			})
			It("generates a password with only uppercase letters", func() {
				pr := types.PasswordRequest{
					Name: "some_pass",
					Type: types.PasswordType,
					Parameters: types.PasswordParams{
						Length:         0,
						ExcludeUpper:   false,
						ExcludeLower:   true,
						ExcludeNumber:  true,
						IncludeSpecial: false,
					},
				}
				// generate 100 passwords
				for i := 0; i < 100; i++ {
					cr, err := pr.Generate(&store.SimpleStore{})
					Expect(err).ToNot(HaveOccurred())
					Expect(cr.(types.PasswordRecord)).ToNot(BeEmpty())
					for _, char := range cr.(types.PasswordRecord) {
						Expect(unicode.IsUpper(char)).To(BeTrue())
					}
				}
			})
			It("generates a password with no digits", func() {
				pr := types.PasswordRequest{
					Name: "some_pass",
					Type: types.PasswordType,
					Parameters: types.PasswordParams{
						Length:         0,
						ExcludeUpper:   false,
						ExcludeLower:   false,
						ExcludeNumber:  true,
						IncludeSpecial: true,
					},
				}
				// generate 100 passwords
				for i := 0; i < 100; i++ {
					cr, err := pr.Generate(&store.SimpleStore{})
					Expect(err).ToNot(HaveOccurred())
					Expect(cr.(types.PasswordRecord)).ToNot(BeEmpty())
					for _, char := range cr.(types.PasswordRecord) {
						Expect(unicode.IsNumber(char)).To(BeFalse())
					}
				}
			})
			It("generates a password with only digits", func() {
				pr := types.PasswordRequest{
					Name: "some_pass",
					Type: types.PasswordType,
					Parameters: types.PasswordParams{
						Length:         0,
						ExcludeUpper:   true,
						ExcludeLower:   true,
						ExcludeNumber:  false,
						IncludeSpecial: false,
					},
				}
				// generate 100 passwords
				for i := 0; i < 100; i++ {
					cr, err := pr.Generate(&store.SimpleStore{})
					Expect(err).ToNot(HaveOccurred())
					Expect(cr.(types.PasswordRecord)).ToNot(BeEmpty())
					for _, char := range cr.(types.PasswordRecord) {
						Expect(unicode.IsNumber(char)).To(BeTrue())
					}
				}
			})
			It("generates a password with no special characters", func() {
				pr := types.PasswordRequest{
					Name: "some_pass",
					Type: types.PasswordType,
					Parameters: types.PasswordParams{
						Length:         0,
						ExcludeUpper:   false,
						ExcludeLower:   false,
						ExcludeNumber:  false,
						IncludeSpecial: false,
					},
				}
				// generate 100 passwords
				for i := 0; i < 100; i++ {
					cr, err := pr.Generate(&store.SimpleStore{})
					Expect(err).ToNot(HaveOccurred())
					Expect(cr.(types.PasswordRecord)).ToNot(BeEmpty())
					for _, char := range cr.(types.PasswordRecord) {
						Expect(unicode.IsNumber(char) || unicode.IsUpper(char) || unicode.IsLower(char)).To(BeTrue())
					}
				}
			})
			It("generates a password with only special characters", func() {
				pr := types.PasswordRequest{
					Name: "some_pass",
					Type: types.PasswordType,
					Parameters: types.PasswordParams{
						Length:         0,
						ExcludeUpper:   true,
						ExcludeLower:   true,
						ExcludeNumber:  true,
						IncludeSpecial: true,
					},
				}
				// generate 100 passwords
				for i := 0; i < 100; i++ {
					cr, err := pr.Generate(&store.SimpleStore{})
					Expect(err).ToNot(HaveOccurred())
					Expect(cr.(types.PasswordRecord)).ToNot(BeEmpty())
					for _, char := range cr.(types.PasswordRecord) {
						Expect(!unicode.IsNumber(char) && !unicode.IsUpper(char) && !unicode.IsLower(char)).To(BeTrue())
					}
				}
			})
		})
	})
})
