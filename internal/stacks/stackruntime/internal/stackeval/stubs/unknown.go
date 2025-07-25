// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package stubs

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/terraform/internal/lang/ephemeral"
	"github.com/hashicorp/terraform/internal/moduletest/mocking"
	"github.com/hashicorp/terraform/internal/providers"
	"github.com/hashicorp/terraform/internal/tfdiags"
)

var _ providers.Interface = (*unknownProvider)(nil)

// unknownProvider is a stub provider that represents a provider that is
// unknown to the current Terraform configuration. This is used when a reference
// to a provider is unknown, or the provider itself has unknown instances.
//
// An unknownProvider is only returned in the context of a provider that should
// have been configured by Stacks. This provider should not be configured again,
// or used for any dedicated offline functionality (such as moving resources and
// provider functions).
type unknownProvider struct {
	unconfiguredClient providers.Interface
}

func UnknownProvider(unconfiguredClient providers.Interface) providers.Interface {
	return &unknownProvider{
		unconfiguredClient: unconfiguredClient,
	}
}

func (u *unknownProvider) GetProviderSchema() providers.GetProviderSchemaResponse {
	// This is offline functionality, so we can hand it off to the unconfigured
	// client.
	return u.unconfiguredClient.GetProviderSchema()
}

func (u *unknownProvider) GetResourceIdentitySchemas() providers.GetResourceIdentitySchemasResponse {
	return u.unconfiguredClient.GetResourceIdentitySchemas()
}

func (u *unknownProvider) ValidateProviderConfig(request providers.ValidateProviderConfigRequest) providers.ValidateProviderConfigResponse {
	// This is offline functionality, so we can hand it off to the unconfigured
	// client.
	return u.unconfiguredClient.ValidateProviderConfig(request)
}

func (u *unknownProvider) ValidateResourceConfig(request providers.ValidateResourceConfigRequest) providers.ValidateResourceConfigResponse {
	// This is offline functionality, so we can hand it off to the unconfigured
	// client.
	return u.unconfiguredClient.ValidateResourceConfig(request)
}

func (u *unknownProvider) ValidateDataResourceConfig(request providers.ValidateDataResourceConfigRequest) providers.ValidateDataResourceConfigResponse {
	// This is offline functionality, so we can hand it off to the unconfigured
	// client.
	return u.unconfiguredClient.ValidateDataResourceConfig(request)
}

func (u *unknownProvider) ValidateListResourceConfig(request providers.ValidateListResourceConfigRequest) providers.ValidateListResourceConfigResponse {
	// This is offline functionality, so we can hand it off to the unconfigured
	// client.
	return u.unconfiguredClient.ValidateListResourceConfig(request)
}

// ValidateEphemeralResourceConfig implements providers.Interface.
func (p *unknownProvider) ValidateEphemeralResourceConfig(providers.ValidateEphemeralResourceConfigRequest) providers.ValidateEphemeralResourceConfigResponse {
	return providers.ValidateEphemeralResourceConfigResponse{
		Diagnostics: nil,
	}
}

func (u *unknownProvider) UpgradeResourceState(request providers.UpgradeResourceStateRequest) providers.UpgradeResourceStateResponse {
	// This is offline functionality, so we can hand it off to the unconfigured
	// client.
	return u.unconfiguredClient.UpgradeResourceState(request)
}

func (u *unknownProvider) UpgradeResourceIdentity(request providers.UpgradeResourceIdentityRequest) providers.UpgradeResourceIdentityResponse {
	// This is offline functionality, so we can hand it off to the unconfigured
	// client.
	return u.unconfiguredClient.UpgradeResourceIdentity(request)
}

func (u *unknownProvider) ConfigureProvider(_ providers.ConfigureProviderRequest) providers.ConfigureProviderResponse {
	// This shouldn't be called, we don't configure an unknown provider within
	// stacks and Terraform Core shouldn't call this method.
	panic("attempted to configure an unknown provider")
}

func (u *unknownProvider) Stop() error {
	// the underlying unconfiguredClient is managed elsewhere.
	return nil
}

func (u *unknownProvider) ReadResource(request providers.ReadResourceRequest) providers.ReadResourceResponse {
	if request.ClientCapabilities.DeferralAllowed {
		// For ReadResource, we'll just return the existing state and defer
		// the operation.
		return providers.ReadResourceResponse{
			NewState: request.PriorState,
			Deferred: &providers.Deferred{
				Reason: providers.DeferredReasonProviderConfigUnknown,
			},
		}
	}
	return providers.ReadResourceResponse{
		Diagnostics: []tfdiags.Diagnostic{
			tfdiags.AttributeValue(
				tfdiags.Error,
				"Provider configuration is unknown",
				"Cannot read from this data source because its associated provider configuration is unknown.",
				nil, // nil attribute path means the overall configuration block
			),
		},
	}
}

func (u *unknownProvider) PlanResourceChange(request providers.PlanResourceChangeRequest) providers.PlanResourceChangeResponse {
	if request.ClientCapabilities.DeferralAllowed {
		// For PlanResourceChange, we'll kind of abuse the mocking library to
		// populate the computed values with unknown values so that future
		// operations can still be used.
		//
		// PlanComputedValuesForResource populates the computed values with
		// unknown values. This isn't the original use case for the mocking
		// library, but it is doing exactly what we need it to do.

		schema := u.GetProviderSchema().ResourceTypes[request.TypeName]
		val, diags := mocking.PlanComputedValuesForResource(request.ProposedNewState, nil, schema.Body)
		if diags.HasErrors() {
			// All the potential errors we get back from this function are
			// related to the user badly defining mocks. We should never hit
			// this as we are just using the default behaviour.
			panic(diags.Err())
		}

		return providers.PlanResourceChangeResponse{
			PlannedState: ephemeral.StripWriteOnlyAttributes(val, schema.Body),
			Deferred: &providers.Deferred{
				Reason: providers.DeferredReasonProviderConfigUnknown,
			},
		}
	}
	return providers.PlanResourceChangeResponse{
		Diagnostics: []tfdiags.Diagnostic{
			tfdiags.AttributeValue(
				tfdiags.Error,
				"Provider configuration is unknown",
				"Cannot plan changes for this resource because its associated provider configuration is unknown.",
				nil, // nil attribute path means the overall configuration block
			),
		},
	}
}

func (u *unknownProvider) ApplyResourceChange(_ providers.ApplyResourceChangeRequest) providers.ApplyResourceChangeResponse {
	return providers.ApplyResourceChangeResponse{
		Diagnostics: []tfdiags.Diagnostic{
			tfdiags.AttributeValue(
				tfdiags.Error,
				"Provider configuration is unknown",
				"Cannot apply changes for this resource because its associated provider configuration is unknown.",
				nil, // nil attribute path means the overall configuration block
			),
		},
	}
}

func (u *unknownProvider) ImportResourceState(request providers.ImportResourceStateRequest) providers.ImportResourceStateResponse {
	if request.ClientCapabilities.DeferralAllowed {
		// For ImportResourceState, we don't have any config to work with and
		// we don't know enough to work out which value the ID corresponds to.
		//
		// We'll just return an unknown value that corresponds to the correct
		// type. Terraform should know how to handle this when it arrives
		// alongside the deferred metadata.

		schema := u.GetProviderSchema().ResourceTypes[request.TypeName]
		return providers.ImportResourceStateResponse{
			ImportedResources: []providers.ImportedResource{
				{
					TypeName: request.TypeName,
					State:    cty.UnknownVal(schema.Body.ImpliedType()),
				},
			},
			Deferred: &providers.Deferred{
				Reason: providers.DeferredReasonProviderConfigUnknown,
			},
		}
	}
	return providers.ImportResourceStateResponse{
		Diagnostics: []tfdiags.Diagnostic{
			tfdiags.AttributeValue(
				tfdiags.Error,
				"Provider configuration is unknown",
				"Cannot import an existing object into this resource because its associated provider configuration is unknown.",
				nil, // nil attribute path means the overall configuration block
			),
		},
	}
}

func (u *unknownProvider) MoveResourceState(_ providers.MoveResourceStateRequest) providers.MoveResourceStateResponse {
	var diags tfdiags.Diagnostics
	diags = diags.Append(tfdiags.AttributeValue(
		tfdiags.Error,
		"Called MoveResourceState on an unknown provider",
		"Terraform called MoveResourceState on an unknown provider. This is a bug in Terraform - please report this error.",
		nil, // nil attribute path means the overall configuration block
	))
	return providers.MoveResourceStateResponse{
		Diagnostics: diags,
	}
}

func (u *unknownProvider) ReadDataSource(request providers.ReadDataSourceRequest) providers.ReadDataSourceResponse {
	if request.ClientCapabilities.DeferralAllowed {
		// For ReadDataSource, we'll kind of abuse the mocking library to
		// populate the computed values with unknown values so that future
		// operations can still be used.
		//
		// PlanComputedValuesForResource populates the computed values with
		// unknown values. This isn't the original use case for the mocking
		// library, but it is doing exactly what we need it to do.

		schema := u.GetProviderSchema().DataSources[request.TypeName]
		val, diags := mocking.PlanComputedValuesForResource(request.Config, nil, schema.Body)
		if diags.HasErrors() {
			// All the potential errors we get back from this function are
			// related to the user badly defining mocks. We should never hit
			// this as we are just using the default behaviour.
			panic(diags.Err())
		}

		return providers.ReadDataSourceResponse{
			State: ephemeral.StripWriteOnlyAttributes(val, schema.Body),
			Deferred: &providers.Deferred{
				Reason: providers.DeferredReasonProviderConfigUnknown,
			},
		}
	}
	return providers.ReadDataSourceResponse{
		Diagnostics: []tfdiags.Diagnostic{
			tfdiags.AttributeValue(
				tfdiags.Error,
				"Provider configuration is unknown",
				"Cannot read from this data source because its associated provider configuration is unknown.",
				nil, // nil attribute path means the overall configuration block
			),
		},
	}
}

// OpenEphemeralResource implements providers.Interface.
func (u *unknownProvider) OpenEphemeralResource(providers.OpenEphemeralResourceRequest) providers.OpenEphemeralResourceResponse {
	// TODO: Once there's a definition for how deferred actions ought to work
	// for ephemeral resource instances, make this report that this one needs
	// to be deferred if the client announced that it supports deferral.
	//
	// For now this is just always an error, because ephemeral resources are
	// just a prototype being developed concurrently with deferred actions.
	var diags tfdiags.Diagnostics
	diags = diags.Append(tfdiags.AttributeValue(
		tfdiags.Error,
		"Provider configuration is unknown",
		"Cannot open this resource instance because its associated provider configuration is unknown.",
		nil, // nil attribute path means the overall configuration block
	))
	return providers.OpenEphemeralResourceResponse{
		Diagnostics: diags,
	}
}

// RenewEphemeralResource implements providers.Interface.
func (u *unknownProvider) RenewEphemeralResource(providers.RenewEphemeralResourceRequest) providers.RenewEphemeralResourceResponse {
	// We don't have anything to do here because OpenEphemeralResource didn't really
	// actually "open" anything.
	return providers.RenewEphemeralResourceResponse{}
}

// CloseEphemeralResource implements providers.Interface.
func (u *unknownProvider) CloseEphemeralResource(providers.CloseEphemeralResourceRequest) providers.CloseEphemeralResourceResponse {
	// We don't have anything to do here because OpenEphemeralResource didn't really
	// actually "open" anything.
	return providers.CloseEphemeralResourceResponse{}
}

func (u *unknownProvider) CallFunction(_ providers.CallFunctionRequest) providers.CallFunctionResponse {
	return providers.CallFunctionResponse{
		Err: fmt.Errorf("CallFunction shouldn't be called on an unknown provider; this is a bug in Terraform - please report this error"),
	}
}

func (u *unknownProvider) ListResource(providers.ListResourceRequest) providers.ListResourceResponse {
	var resp providers.ListResourceResponse
	resp.Diagnostics = resp.Diagnostics.Append(tfdiags.AttributeValue(
		tfdiags.Error,
		"Called ListResource on an unknown provider",
		"Terraform called ListResource on an unknown provider. This is a bug in Terraform - please report this error.",
		nil, // nil attribute path means the overall configuration block
	))
	return resp
}

// ValidateStateStoreConfig implements providers.Interface.
func (u *unknownProvider) ValidateStateStoreConfig(providers.ValidateStateStoreConfigRequest) providers.ValidateStateStoreConfigResponse {
	var diags tfdiags.Diagnostics
	diags = diags.Append(tfdiags.AttributeValue(
		tfdiags.Error,
		"Provider configuration is unknown",
		"Cannot validate this state store because its associated provider configuration is unknown.",
		nil, // nil attribute path means the overall configuration block
	))
	return providers.ValidateStateStoreConfigResponse{
		Diagnostics: diags,
	}
}

// ConfigureStateStore implements providers.Interface.
func (u *unknownProvider) ConfigureStateStore(providers.ConfigureStateStoreRequest) providers.ConfigureStateStoreResponse {
	var diags tfdiags.Diagnostics
	diags = diags.Append(tfdiags.AttributeValue(
		tfdiags.Error,
		"Provider configuration is unknown",
		"Cannot configure this state store because its associated provider configuration is unknown.",
		nil, // nil attribute path means the overall configuration block
	))
	return providers.ConfigureStateStoreResponse{
		Diagnostics: diags,
	}
}

// GetStates implements providers.Interface.
func (u *unknownProvider) GetStates(providers.GetStatesRequest) providers.GetStatesResponse {
	var diags tfdiags.Diagnostics
	diags = diags.Append(tfdiags.AttributeValue(
		tfdiags.Error,
		"Provider configuration is unknown",
		"Cannot list states managed by this state store because its associated provider configuration is unknown.",
		nil, // nil attribute path means the overall configuration block
	))
	return providers.GetStatesResponse{
		Diagnostics: diags,
	}
}

// DeleteState implements providers.Interface.
func (u *unknownProvider) DeleteState(providers.DeleteStateRequest) providers.DeleteStateResponse {
	var diags tfdiags.Diagnostics
	diags = diags.Append(tfdiags.AttributeValue(
		tfdiags.Error,
		"Provider configuration is unknown",
		"Cannot use this state store to delete a state because its associated provider configuration is unknown.",
		nil, // nil attribute path means the overall configuration block
	))
	return providers.DeleteStateResponse{
		Diagnostics: diags,
	}
}

// PlanAction implements providers.Interface.
func (u *unknownProvider) PlanAction(request providers.PlanActionRequest) providers.PlanActionResponse {
	// TODO: Once actions support deferrals we can implement this
	return providers.PlanActionResponse{
		Diagnostics: []tfdiags.Diagnostic{
			tfdiags.AttributeValue(
				tfdiags.Error,
				"Provider configuration is unknown",
				"Cannot plan this action because its associated provider configuration is unknown.",
				nil, // nil attribute path means the overall configuration block
			),
		},
	}
}

// InvokeAction implements providers.Interface.
func (u *unknownProvider) InvokeAction(request providers.InvokeActionRequest) providers.InvokeActionResponse {
	return providers.InvokeActionResponse{
		Diagnostics: []tfdiags.Diagnostic{
			tfdiags.AttributeValue(
				tfdiags.Error,
				"Provider configuration is unknown",
				"Cannot invoke this action because its associated provider configuration is unknown.",
				nil, // nil attribute path means the overall configuration block
			),
		},
	}
}

func (u *unknownProvider) ValidateActionConfig(request providers.ValidateActionConfigRequest) providers.ValidateActionConfigResponse {
	return providers.ValidateActionConfigResponse{
		Diagnostics: []tfdiags.Diagnostic{
			tfdiags.AttributeValue(
				tfdiags.Error,
				"Provider configuration is unknown",
				"Cannot validate this action configuration because its associated provider configuration is unknown.",
				nil, // nil attribute path means the overall configuration block
			),
		},
	}
}

func (u *unknownProvider) Close() error {
	// the underlying unconfiguredClient is managed elsewhere.
	return nil
}
