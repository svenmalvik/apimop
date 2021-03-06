/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "no.vipps/guestbook/api/v1"
)

// GuestbookReconciler reconciles a Guestbook object
type GuestbookReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	subscriptionID    string
	location          = "westeurope"
	resourceGroupName = "smatestbla2-rg"
)

//+kubebuilder:rbac:groups=webapp.no.vipps,resources=guestbooks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.no.vipps,resources=guestbooks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.no.vipps,resources=guestbooks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Guestbook object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *GuestbookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	// TODO(user): your logic here
	//log.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	l.Info("Bla-------------------")

	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		l.Info("AZURE_SUBSCRIPTION_ID is not set")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		l.Error(err, "azidentity err")
	}
	ctxx := context.Background()
	var guestbook webappv1.Guestbook
	r.Get(ctxx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, &guestbook)

	createResourceGroup(ctxx, cred, guestbook.Spec.Name)
	if err != nil {
		l.Error(err, "createResourceGroup")
	}

	createAPI(ctxx)
	//createPureAPI(ctxx)
	//getAPI(ctxx)

	return ctrl.Result{}, nil
}

func createPureAPI(ctx context.Context) {
	l := log.FromContext(ctx)
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		l.Error(err, "failed to obtain a credential: %v")
	}

	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	client := armapimanagement.NewAPIClient(subscriptionID, cred, nil)
	poller, err := client.BeginCreateOrUpdate(ctx,
		"smaapim-rg",
		"smaapim2",
		"demo-conference-api-3",
		armapimanagement.APICreateOrUpdateParameter{
			Properties: &armapimanagement.APICreateOrUpdateProperties{
				Description: to.StringPtr("Dies ist ein Test"),
				//AuthenticationSettings: &armapimanagement.AuthenticationSettingsContract{
				//	OAuth2: &armapimanagement.OAuth2AuthenticationSettingsContract{
				//		AuthorizationServerID: to.StringPtr("<authorization-server-id>"),
				//		Scope:                 to.StringPtr("<scope>"),
				//	},
				//},
				SubscriptionKeyParameterNames: &armapimanagement.SubscriptionKeyParameterNamesContract{
					Header: to.StringPtr("X-ApiKey"),
					Query:  to.StringPtr("ApiKey"),
				},
				Path:        to.StringPtr("conf"),
				DisplayName: to.StringPtr("democonf3"),
				Protocols: []*armapimanagement.Protocol{
					armapimanagement.Protocol("https").ToPtr(),
					armapimanagement.Protocol("http").ToPtr()},
				ServiceURL: to.StringPtr("https://svenmalvik.com"),
			},
		},
		&armapimanagement.APIClientBeginCreateOrUpdateOptions{IfMatch: nil})
	if err != nil {
		l.Error(err, "Error 1")
	}
	res, err := poller.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		l.Error(err, "Error 2")
	}
	l.Info("Response result: %#v\n", res.APIClientCreateOrUpdateResult)
}

func getAPI(ctx context.Context) {
	l := log.FromContext(ctx)
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		l.Error(err, "failed to obtain a credential")
	}
	//ctxx := context.Background()
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	client := armapimanagement.NewAPIClient(subscriptionID, cred, nil)
	res, err := client.Get(ctx,
		"smaapim-rg",
		"smaapim2",
		"demo-conference-api",
		nil)
	if err != nil {
		l.Error(err, "----")
	}
	l.Info("Response result: %#v\n", res.APIClientGetResult)
}

func createResourceGroup(ctxx context.Context, cred azcore.TokenCredential, rg string) (*armresources.ResourceGroup, error) {
	l := log.FromContext(ctxx)
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

	l.Info(rg)
	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctxx,
		rg,
		armresources.ResourceGroup{
			Location: to.StringPtr(location),
		},
		nil)

	if err != nil {
		return nil, err
	}

	return &resourceGroupResp.ResourceGroup, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GuestbookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.Guestbook{}).
		Complete(r)
}
