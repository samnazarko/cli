package cmd

import (
	"errors"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/ukfast/cli/test/mocks"
	"github.com/ukfast/cli/test/test_output"
	"github.com/ukfast/sdk-go/pkg/service/account"
)

func Test_accountInvoiceList(t *testing.T) {
	t.Run("DefaultRetrieve", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		service := mocks.NewMockAccountService(mockCtrl)

		service.EXPECT().GetInvoices(gomock.Any()).Return([]account.Invoice{}, nil).Times(1)

		accountInvoiceList(service, &cobra.Command{}, []string{})
	})

	t.Run("MalformedFlag_OutputsFatal", func(t *testing.T) {
		defer func() { flagFilter = nil }()

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		service := mocks.NewMockAccountService(mockCtrl)
		flagFilter = []string{"invalidfilter"}

		test_output.AssertFatalOutput(t, "Missing value for filtering\n", func() {
			accountInvoiceList(service, &cobra.Command{}, []string{})
		})
	})

	t.Run("GetInvoicesError_OutputsFatal", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		service := mocks.NewMockAccountService(mockCtrl)

		service.EXPECT().GetInvoices(gomock.Any()).Return([]account.Invoice{}, errors.New("test error")).Times(1)

		test_output.AssertFatalOutput(t, "Error retrieving invoices: test error\n", func() {
			accountInvoiceList(service, &cobra.Command{}, []string{})
		})
	})
}

func Test_accountInvoiceShowCmd_Args(t *testing.T) {
	t.Run("ValidArgs_NoError", func(t *testing.T) {
		err := accountInvoiceShowCmd().Args(nil, []string{"123"})

		assert.Nil(t, err)
	})

	t.Run("InvalidArgs_Error", func(t *testing.T) {
		err := accountInvoiceShowCmd().Args(nil, []string{})

		assert.NotNil(t, err)
		assert.Equal(t, "Missing invoice", err.Error())
	})
}

func Test_accountInvoiceShow(t *testing.T) {
	t.Run("SingleInvoice", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		service := mocks.NewMockAccountService(mockCtrl)

		service.EXPECT().GetInvoice(123).Return(account.Invoice{}, nil).Times(1)

		accountInvoiceShow(service, &cobra.Command{}, []string{"123"})
	})

	t.Run("MultipleInvoices", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		service := mocks.NewMockAccountService(mockCtrl)

		gomock.InOrder(
			service.EXPECT().GetInvoice(123).Return(account.Invoice{}, nil),
			service.EXPECT().GetInvoice(456).Return(account.Invoice{}, nil),
		)

		accountInvoiceShow(service, &cobra.Command{}, []string{"123", "456"})
	})

	t.Run("GetInvoiceID_OutputsError", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		service := mocks.NewMockAccountService(mockCtrl)

		test_output.AssertErrorOutput(t, "Invalid invoice ID [abc]\n", func() {
			accountInvoiceShow(service, &cobra.Command{}, []string{"abc"})
		})
	})

	t.Run("GetInvoiceError_OutputsError", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		service := mocks.NewMockAccountService(mockCtrl)

		service.EXPECT().GetInvoice(123).Return(account.Invoice{}, errors.New("test error"))

		test_output.AssertErrorOutput(t, "Error retrieving invoice [123]: test error\n", func() {
			accountInvoiceShow(service, &cobra.Command{}, []string{"123"})
		})
	})
}