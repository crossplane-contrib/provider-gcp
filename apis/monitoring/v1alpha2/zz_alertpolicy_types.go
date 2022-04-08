/*
Copyright 2019 The Crossplane Authors.

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

// Code generated by terrajet. DO NOT EDIT.

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

type AggregationsObservation struct {
}

type AggregationsParameters struct {

	// The alignment period for per-time
	// series alignment. If present,
	// alignmentPeriod must be at least
	// 60 seconds. After per-time series
	// alignment, each time series will
	// contain data points only on the
	// period boundaries. If
	// perSeriesAligner is not specified
	// or equals ALIGN_NONE, then this
	// field is ignored. If
	// perSeriesAligner is specified and
	// does not equal ALIGN_NONE, then
	// this field must be defined;
	// otherwise an error is returned.
	// +kubebuilder:validation:Optional
	AlignmentPeriod *string `json:"alignmentPeriod,omitempty" tf:"alignment_period,omitempty"`

	// The approach to be used to combine
	// time series. Not all reducer
	// functions may be applied to all
	// time series, depending on the
	// metric type and the value type of
	// the original time series.
	// Reduction may change the metric
	// type of value type of the time
	// series.Time series data must be
	// aligned in order to perform cross-
	// time series reduction. If
	// crossSeriesReducer is specified,
	// then perSeriesAligner must be
	// specified and not equal ALIGN_NONE
	// and alignmentPeriod must be
	// specified; otherwise, an error is
	// returned. Possible values: ["REDUCE_NONE", "REDUCE_MEAN", "REDUCE_MIN", "REDUCE_MAX", "REDUCE_SUM", "REDUCE_STDDEV", "REDUCE_COUNT", "REDUCE_COUNT_TRUE", "REDUCE_COUNT_FALSE", "REDUCE_FRACTION_TRUE", "REDUCE_PERCENTILE_99", "REDUCE_PERCENTILE_95", "REDUCE_PERCENTILE_50", "REDUCE_PERCENTILE_05"]
	// +kubebuilder:validation:Optional
	CrossSeriesReducer *string `json:"crossSeriesReducer,omitempty" tf:"cross_series_reducer,omitempty"`

	// The set of fields to preserve when
	// crossSeriesReducer is specified.
	// The groupByFields determine how
	// the time series are partitioned
	// into subsets prior to applying the
	// aggregation function. Each subset
	// contains time series that have the
	// same value for each of the
	// grouping fields. Each individual
	// time series is a member of exactly
	// one subset. The crossSeriesReducer
	// is applied to each subset of time
	// series. It is not possible to
	// reduce across different resource
	// types, so this field implicitly
	// contains resource.type. Fields not
	// specified in groupByFields are
	// aggregated away. If groupByFields
	// is not specified and all the time
	// series have the same resource
	// type, then the time series are
	// aggregated into a single output
	// time series. If crossSeriesReducer
	// is not defined, this field is
	// ignored.
	// +kubebuilder:validation:Optional
	GroupByFields []*string `json:"groupByFields,omitempty" tf:"group_by_fields,omitempty"`

	// The approach to be used to align
	// individual time series. Not all
	// alignment functions may be applied
	// to all time series, depending on
	// the metric type and value type of
	// the original time series.
	// Alignment may change the metric
	// type or the value type of the time
	// series.Time series data must be
	// aligned in order to perform cross-
	// time series reduction. If
	// crossSeriesReducer is specified,
	// then perSeriesAligner must be
	// specified and not equal ALIGN_NONE
	// and alignmentPeriod must be
	// specified; otherwise, an error is
	// returned. Possible values: ["ALIGN_NONE", "ALIGN_DELTA", "ALIGN_RATE", "ALIGN_INTERPOLATE", "ALIGN_NEXT_OLDER", "ALIGN_MIN", "ALIGN_MAX", "ALIGN_MEAN", "ALIGN_COUNT", "ALIGN_SUM", "ALIGN_STDDEV", "ALIGN_COUNT_TRUE", "ALIGN_COUNT_FALSE", "ALIGN_FRACTION_TRUE", "ALIGN_PERCENTILE_99", "ALIGN_PERCENTILE_95", "ALIGN_PERCENTILE_50", "ALIGN_PERCENTILE_05", "ALIGN_PERCENT_CHANGE"]
	// +kubebuilder:validation:Optional
	PerSeriesAligner *string `json:"perSeriesAligner,omitempty" tf:"per_series_aligner,omitempty"`
}

type AlertPolicyObservation struct {
	CreationRecord []CreationRecordObservation `json:"creationRecord,omitempty" tf:"creation_record,omitempty"`

	ID *string `json:"id,omitempty" tf:"id,omitempty"`

	Name *string `json:"name,omitempty" tf:"name,omitempty"`
}

type AlertPolicyParameters struct {

	// How to combine the results of multiple conditions to
	// determine if an incident should be opened. Possible values: ["AND", "OR", "AND_WITH_MATCHING_RESOURCE"]
	// +kubebuilder:validation:Required
	Combiner *string `json:"combiner" tf:"combiner,omitempty"`

	// A list of conditions for the policy. The conditions are combined by
	// AND or OR according to the combiner field. If the combined conditions
	// evaluate to true, then an incident is created. A policy can have from
	// one to six conditions.
	// +kubebuilder:validation:Required
	Conditions []ConditionsParameters `json:"conditions" tf:"conditions,omitempty"`

	// A short name or phrase used to identify the policy in
	// dashboards, notifications, and incidents. To avoid confusion, don't use
	// the same display name for multiple policies in the same project. The
	// name is limited to 512 Unicode characters.
	// +kubebuilder:validation:Required
	DisplayName *string `json:"displayName" tf:"display_name,omitempty"`

	// Documentation that is included with notifications and incidents related
	// to this policy. Best practice is for the documentation to include information
	// to help responders understand, mitigate, escalate, and correct the underlying
	// problems detected by the alerting policy. Notification channels that have
	// limited capacity might not show this documentation.
	// +kubebuilder:validation:Optional
	Documentation []DocumentationParameters `json:"documentation,omitempty" tf:"documentation,omitempty"`

	// Whether or not the policy is enabled. The default is true.
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty" tf:"enabled,omitempty"`

	// Identifies the notification channels to which notifications should be
	// sent when incidents are opened or closed or when new violations occur
	// on an already opened incident. Each element of this array corresponds
	// to the name field in each of the NotificationChannel objects that are
	// returned from the notificationChannels.list method. The syntax of the
	// entries in this field is
	// 'projects/[PROJECT_ID]/notificationChannels/[CHANNEL_ID]'
	// +kubebuilder:validation:Optional
	NotificationChannels []*string `json:"notificationChannels,omitempty" tf:"notification_channels,omitempty"`

	// +kubebuilder:validation:Optional
	Project *string `json:"project,omitempty" tf:"project,omitempty"`

	// This field is intended to be used for organizing and identifying the AlertPolicy
	// objects.The field can contain up to 64 entries. Each key and value is limited
	// to 63 Unicode characters or 128 bytes, whichever is smaller. Labels and values
	// can contain only lowercase letters, numerals, underscores, and dashes. Keys
	// must begin with a letter.
	// +kubebuilder:validation:Optional
	UserLabels map[string]*string `json:"userLabels,omitempty" tf:"user_labels,omitempty"`
}

type ConditionAbsentObservation struct {
}

type ConditionAbsentParameters struct {

	// Specifies the alignment of data points in
	// individual time series as well as how to
	// combine the retrieved time series together
	// (such as when aggregating multiple streams
	// on each resource to a single stream for each
	// resource or when aggregating streams across
	// all members of a group of resources).
	// Multiple aggregations are applied in the
	// order specified.
	// +kubebuilder:validation:Optional
	Aggregations []AggregationsParameters `json:"aggregations,omitempty" tf:"aggregations,omitempty"`

	// The amount of time that a time series must
	// fail to report new data to be considered
	// failing. Currently, only values that are a
	// multiple of a minute--e.g. 60s, 120s, or 300s
	// --are supported.
	// +kubebuilder:validation:Required
	Duration *string `json:"duration" tf:"duration,omitempty"`

	// A filter that identifies which time series
	// should be compared with the threshold.The
	// filter is similar to the one that is
	// specified in the
	// MetricService.ListTimeSeries request (that
	// call is useful to verify the time series
	// that will be retrieved / processed) and must
	// specify the metric type and optionally may
	// contain restrictions on resource type,
	// resource labels, and metric labels. This
	// field may not exceed 2048 Unicode characters
	// in length.
	// +kubebuilder:validation:Optional
	Filter *string `json:"filter,omitempty" tf:"filter,omitempty"`

	// The number/percent of time series for which
	// the comparison must hold in order for the
	// condition to trigger. If unspecified, then
	// the condition will trigger if the comparison
	// is true for any of the time series that have
	// been identified by filter and aggregations.
	// +kubebuilder:validation:Optional
	Trigger []TriggerParameters `json:"trigger,omitempty" tf:"trigger,omitempty"`
}

type ConditionMonitoringQueryLanguageObservation struct {
}

type ConditionMonitoringQueryLanguageParameters struct {

	// The amount of time that a time series must
	// violate the threshold to be considered
	// failing. Currently, only values that are a
	// multiple of a minute--e.g., 0, 60, 120, or
	// 300 seconds--are supported. If an invalid
	// value is given, an error will be returned.
	// When choosing a duration, it is useful to
	// keep in mind the frequency of the underlying
	// time series data (which may also be affected
	// by any alignments specified in the
	// aggregations field); a good duration is long
	// enough so that a single outlier does not
	// generate spurious alerts, but short enough
	// that unhealthy states are detected and
	// alerted on quickly.
	// +kubebuilder:validation:Required
	Duration *string `json:"duration" tf:"duration,omitempty"`

	// Monitoring Query Language query that outputs a boolean stream.
	// +kubebuilder:validation:Required
	Query *string `json:"query" tf:"query,omitempty"`

	// The number/percent of time series for which
	// the comparison must hold in order for the
	// condition to trigger. If unspecified, then
	// the condition will trigger if the comparison
	// is true for any of the time series that have
	// been identified by filter and aggregations,
	// or by the ratio, if denominator_filter and
	// denominator_aggregations are specified.
	// +kubebuilder:validation:Optional
	Trigger []ConditionMonitoringQueryLanguageTriggerParameters `json:"trigger,omitempty" tf:"trigger,omitempty"`
}

type ConditionMonitoringQueryLanguageTriggerObservation struct {
}

type ConditionMonitoringQueryLanguageTriggerParameters struct {

	// The absolute number of time series
	// that must fail the predicate for the
	// condition to be triggered.
	// +kubebuilder:validation:Optional
	Count *float64 `json:"count,omitempty" tf:"count,omitempty"`

	// The percentage of time series that
	// must fail the predicate for the
	// condition to be triggered.
	// +kubebuilder:validation:Optional
	Percent *float64 `json:"percent,omitempty" tf:"percent,omitempty"`
}

type ConditionThresholdAggregationsObservation struct {
}

type ConditionThresholdAggregationsParameters struct {

	// The alignment period for per-time
	// series alignment. If present,
	// alignmentPeriod must be at least
	// 60 seconds. After per-time series
	// alignment, each time series will
	// contain data points only on the
	// period boundaries. If
	// perSeriesAligner is not specified
	// or equals ALIGN_NONE, then this
	// field is ignored. If
	// perSeriesAligner is specified and
	// does not equal ALIGN_NONE, then
	// this field must be defined;
	// otherwise an error is returned.
	// +kubebuilder:validation:Optional
	AlignmentPeriod *string `json:"alignmentPeriod,omitempty" tf:"alignment_period,omitempty"`

	// The approach to be used to combine
	// time series. Not all reducer
	// functions may be applied to all
	// time series, depending on the
	// metric type and the value type of
	// the original time series.
	// Reduction may change the metric
	// type of value type of the time
	// series.Time series data must be
	// aligned in order to perform cross-
	// time series reduction. If
	// crossSeriesReducer is specified,
	// then perSeriesAligner must be
	// specified and not equal ALIGN_NONE
	// and alignmentPeriod must be
	// specified; otherwise, an error is
	// returned. Possible values: ["REDUCE_NONE", "REDUCE_MEAN", "REDUCE_MIN", "REDUCE_MAX", "REDUCE_SUM", "REDUCE_STDDEV", "REDUCE_COUNT", "REDUCE_COUNT_TRUE", "REDUCE_COUNT_FALSE", "REDUCE_FRACTION_TRUE", "REDUCE_PERCENTILE_99", "REDUCE_PERCENTILE_95", "REDUCE_PERCENTILE_50", "REDUCE_PERCENTILE_05"]
	// +kubebuilder:validation:Optional
	CrossSeriesReducer *string `json:"crossSeriesReducer,omitempty" tf:"cross_series_reducer,omitempty"`

	// The set of fields to preserve when
	// crossSeriesReducer is specified.
	// The groupByFields determine how
	// the time series are partitioned
	// into subsets prior to applying the
	// aggregation function. Each subset
	// contains time series that have the
	// same value for each of the
	// grouping fields. Each individual
	// time series is a member of exactly
	// one subset. The crossSeriesReducer
	// is applied to each subset of time
	// series. It is not possible to
	// reduce across different resource
	// types, so this field implicitly
	// contains resource.type. Fields not
	// specified in groupByFields are
	// aggregated away. If groupByFields
	// is not specified and all the time
	// series have the same resource
	// type, then the time series are
	// aggregated into a single output
	// time series. If crossSeriesReducer
	// is not defined, this field is
	// ignored.
	// +kubebuilder:validation:Optional
	GroupByFields []*string `json:"groupByFields,omitempty" tf:"group_by_fields,omitempty"`

	// The approach to be used to align
	// individual time series. Not all
	// alignment functions may be applied
	// to all time series, depending on
	// the metric type and value type of
	// the original time series.
	// Alignment may change the metric
	// type or the value type of the time
	// series.Time series data must be
	// aligned in order to perform cross-
	// time series reduction. If
	// crossSeriesReducer is specified,
	// then perSeriesAligner must be
	// specified and not equal ALIGN_NONE
	// and alignmentPeriod must be
	// specified; otherwise, an error is
	// returned. Possible values: ["ALIGN_NONE", "ALIGN_DELTA", "ALIGN_RATE", "ALIGN_INTERPOLATE", "ALIGN_NEXT_OLDER", "ALIGN_MIN", "ALIGN_MAX", "ALIGN_MEAN", "ALIGN_COUNT", "ALIGN_SUM", "ALIGN_STDDEV", "ALIGN_COUNT_TRUE", "ALIGN_COUNT_FALSE", "ALIGN_FRACTION_TRUE", "ALIGN_PERCENTILE_99", "ALIGN_PERCENTILE_95", "ALIGN_PERCENTILE_50", "ALIGN_PERCENTILE_05", "ALIGN_PERCENT_CHANGE"]
	// +kubebuilder:validation:Optional
	PerSeriesAligner *string `json:"perSeriesAligner,omitempty" tf:"per_series_aligner,omitempty"`
}

type ConditionThresholdObservation struct {
}

type ConditionThresholdParameters struct {

	// Specifies the alignment of data points in
	// individual time series as well as how to
	// combine the retrieved time series together
	// (such as when aggregating multiple streams
	// on each resource to a single stream for each
	// resource or when aggregating streams across
	// all members of a group of resources).
	// Multiple aggregations are applied in the
	// order specified.This field is similar to the
	// one in the MetricService.ListTimeSeries
	// request. It is advisable to use the
	// ListTimeSeries method when debugging this
	// field.
	// +kubebuilder:validation:Optional
	Aggregations []ConditionThresholdAggregationsParameters `json:"aggregations,omitempty" tf:"aggregations,omitempty"`

	// The comparison to apply between the time
	// series (indicated by filter and aggregation)
	// and the threshold (indicated by
	// threshold_value). The comparison is applied
	// on each time series, with the time series on
	// the left-hand side and the threshold on the
	// right-hand side. Only COMPARISON_LT and
	// COMPARISON_GT are supported currently. Possible values: ["COMPARISON_GT", "COMPARISON_GE", "COMPARISON_LT", "COMPARISON_LE", "COMPARISON_EQ", "COMPARISON_NE"]
	// +kubebuilder:validation:Required
	Comparison *string `json:"comparison" tf:"comparison,omitempty"`

	// Specifies the alignment of data points in
	// individual time series selected by
	// denominatorFilter as well as how to combine
	// the retrieved time series together (such as
	// when aggregating multiple streams on each
	// resource to a single stream for each
	// resource or when aggregating streams across
	// all members of a group of resources).When
	// computing ratios, the aggregations and
	// denominator_aggregations fields must use the
	// same alignment period and produce time
	// series that have the same periodicity and
	// labels.This field is similar to the one in
	// the MetricService.ListTimeSeries request. It
	// is advisable to use the ListTimeSeries
	// method when debugging this field.
	// +kubebuilder:validation:Optional
	DenominatorAggregations []DenominatorAggregationsParameters `json:"denominatorAggregations,omitempty" tf:"denominator_aggregations,omitempty"`

	// A filter that identifies a time series that
	// should be used as the denominator of a ratio
	// that will be compared with the threshold. If
	// a denominator_filter is specified, the time
	// series specified by the filter field will be
	// used as the numerator.The filter is similar
	// to the one that is specified in the
	// MetricService.ListTimeSeries request (that
	// call is useful to verify the time series
	// that will be retrieved / processed) and must
	// specify the metric type and optionally may
	// contain restrictions on resource type,
	// resource labels, and metric labels. This
	// field may not exceed 2048 Unicode characters
	// in length.
	// +kubebuilder:validation:Optional
	DenominatorFilter *string `json:"denominatorFilter,omitempty" tf:"denominator_filter,omitempty"`

	// The amount of time that a time series must
	// violate the threshold to be considered
	// failing. Currently, only values that are a
	// multiple of a minute--e.g., 0, 60, 120, or
	// 300 seconds--are supported. If an invalid
	// value is given, an error will be returned.
	// When choosing a duration, it is useful to
	// keep in mind the frequency of the underlying
	// time series data (which may also be affected
	// by any alignments specified in the
	// aggregations field); a good duration is long
	// enough so that a single outlier does not
	// generate spurious alerts, but short enough
	// that unhealthy states are detected and
	// alerted on quickly.
	// +kubebuilder:validation:Required
	Duration *string `json:"duration" tf:"duration,omitempty"`

	// A filter that identifies which time series
	// should be compared with the threshold.The
	// filter is similar to the one that is
	// specified in the
	// MetricService.ListTimeSeries request (that
	// call is useful to verify the time series
	// that will be retrieved / processed) and must
	// specify the metric type and optionally may
	// contain restrictions on resource type,
	// resource labels, and metric labels. This
	// field may not exceed 2048 Unicode characters
	// in length.
	// +kubebuilder:validation:Optional
	Filter *string `json:"filter,omitempty" tf:"filter,omitempty"`

	// A value against which to compare the time
	// series.
	// +kubebuilder:validation:Optional
	ThresholdValue *float64 `json:"thresholdValue,omitempty" tf:"threshold_value,omitempty"`

	// The number/percent of time series for which
	// the comparison must hold in order for the
	// condition to trigger. If unspecified, then
	// the condition will trigger if the comparison
	// is true for any of the time series that have
	// been identified by filter and aggregations,
	// or by the ratio, if denominator_filter and
	// denominator_aggregations are specified.
	// +kubebuilder:validation:Optional
	Trigger []ConditionThresholdTriggerParameters `json:"trigger,omitempty" tf:"trigger,omitempty"`
}

type ConditionThresholdTriggerObservation struct {
}

type ConditionThresholdTriggerParameters struct {

	// The absolute number of time series
	// that must fail the predicate for the
	// condition to be triggered.
	// +kubebuilder:validation:Optional
	Count *float64 `json:"count,omitempty" tf:"count,omitempty"`

	// The percentage of time series that
	// must fail the predicate for the
	// condition to be triggered.
	// +kubebuilder:validation:Optional
	Percent *float64 `json:"percent,omitempty" tf:"percent,omitempty"`
}

type ConditionsObservation struct {
	Name *string `json:"name,omitempty" tf:"name,omitempty"`
}

type ConditionsParameters struct {

	// A condition that checks that a time series
	// continues to receive new data points.
	// +kubebuilder:validation:Optional
	ConditionAbsent []ConditionAbsentParameters `json:"conditionAbsent,omitempty" tf:"condition_absent,omitempty"`

	// A Monitoring Query Language query that outputs a boolean stream
	// +kubebuilder:validation:Optional
	ConditionMonitoringQueryLanguage []ConditionMonitoringQueryLanguageParameters `json:"conditionMonitoringQueryLanguage,omitempty" tf:"condition_monitoring_query_language,omitempty"`

	// A condition that compares a time series against a
	// threshold.
	// +kubebuilder:validation:Optional
	ConditionThreshold []ConditionThresholdParameters `json:"conditionThreshold,omitempty" tf:"condition_threshold,omitempty"`

	// A short name or phrase used to identify the
	// condition in dashboards, notifications, and
	// incidents. To avoid confusion, don't use the same
	// display name for multiple conditions in the same
	// policy.
	// +kubebuilder:validation:Required
	DisplayName *string `json:"displayName" tf:"display_name,omitempty"`
}

type CreationRecordObservation struct {
	MutateTime *string `json:"mutateTime,omitempty" tf:"mutate_time,omitempty"`

	MutatedBy *string `json:"mutatedBy,omitempty" tf:"mutated_by,omitempty"`
}

type CreationRecordParameters struct {
}

type DenominatorAggregationsObservation struct {
}

type DenominatorAggregationsParameters struct {

	// The alignment period for per-time
	// series alignment. If present,
	// alignmentPeriod must be at least
	// 60 seconds. After per-time series
	// alignment, each time series will
	// contain data points only on the
	// period boundaries. If
	// perSeriesAligner is not specified
	// or equals ALIGN_NONE, then this
	// field is ignored. If
	// perSeriesAligner is specified and
	// does not equal ALIGN_NONE, then
	// this field must be defined;
	// otherwise an error is returned.
	// +kubebuilder:validation:Optional
	AlignmentPeriod *string `json:"alignmentPeriod,omitempty" tf:"alignment_period,omitempty"`

	// The approach to be used to combine
	// time series. Not all reducer
	// functions may be applied to all
	// time series, depending on the
	// metric type and the value type of
	// the original time series.
	// Reduction may change the metric
	// type of value type of the time
	// series.Time series data must be
	// aligned in order to perform cross-
	// time series reduction. If
	// crossSeriesReducer is specified,
	// then perSeriesAligner must be
	// specified and not equal ALIGN_NONE
	// and alignmentPeriod must be
	// specified; otherwise, an error is
	// returned. Possible values: ["REDUCE_NONE", "REDUCE_MEAN", "REDUCE_MIN", "REDUCE_MAX", "REDUCE_SUM", "REDUCE_STDDEV", "REDUCE_COUNT", "REDUCE_COUNT_TRUE", "REDUCE_COUNT_FALSE", "REDUCE_FRACTION_TRUE", "REDUCE_PERCENTILE_99", "REDUCE_PERCENTILE_95", "REDUCE_PERCENTILE_50", "REDUCE_PERCENTILE_05"]
	// +kubebuilder:validation:Optional
	CrossSeriesReducer *string `json:"crossSeriesReducer,omitempty" tf:"cross_series_reducer,omitempty"`

	// The set of fields to preserve when
	// crossSeriesReducer is specified.
	// The groupByFields determine how
	// the time series are partitioned
	// into subsets prior to applying the
	// aggregation function. Each subset
	// contains time series that have the
	// same value for each of the
	// grouping fields. Each individual
	// time series is a member of exactly
	// one subset. The crossSeriesReducer
	// is applied to each subset of time
	// series. It is not possible to
	// reduce across different resource
	// types, so this field implicitly
	// contains resource.type. Fields not
	// specified in groupByFields are
	// aggregated away. If groupByFields
	// is not specified and all the time
	// series have the same resource
	// type, then the time series are
	// aggregated into a single output
	// time series. If crossSeriesReducer
	// is not defined, this field is
	// ignored.
	// +kubebuilder:validation:Optional
	GroupByFields []*string `json:"groupByFields,omitempty" tf:"group_by_fields,omitempty"`

	// The approach to be used to align
	// individual time series. Not all
	// alignment functions may be applied
	// to all time series, depending on
	// the metric type and value type of
	// the original time series.
	// Alignment may change the metric
	// type or the value type of the time
	// series.Time series data must be
	// aligned in order to perform cross-
	// time series reduction. If
	// crossSeriesReducer is specified,
	// then perSeriesAligner must be
	// specified and not equal ALIGN_NONE
	// and alignmentPeriod must be
	// specified; otherwise, an error is
	// returned. Possible values: ["ALIGN_NONE", "ALIGN_DELTA", "ALIGN_RATE", "ALIGN_INTERPOLATE", "ALIGN_NEXT_OLDER", "ALIGN_MIN", "ALIGN_MAX", "ALIGN_MEAN", "ALIGN_COUNT", "ALIGN_SUM", "ALIGN_STDDEV", "ALIGN_COUNT_TRUE", "ALIGN_COUNT_FALSE", "ALIGN_FRACTION_TRUE", "ALIGN_PERCENTILE_99", "ALIGN_PERCENTILE_95", "ALIGN_PERCENTILE_50", "ALIGN_PERCENTILE_05", "ALIGN_PERCENT_CHANGE"]
	// +kubebuilder:validation:Optional
	PerSeriesAligner *string `json:"perSeriesAligner,omitempty" tf:"per_series_aligner,omitempty"`
}

type DocumentationObservation struct {
}

type DocumentationParameters struct {

	// The text of the documentation, interpreted according to mimeType.
	// The content may not exceed 8,192 Unicode characters and may not
	// exceed more than 10,240 bytes when encoded in UTF-8 format,
	// whichever is smaller.
	// +kubebuilder:validation:Optional
	Content *string `json:"content,omitempty" tf:"content,omitempty"`

	// The format of the content field. Presently, only the value
	// "text/markdown" is supported.
	// +kubebuilder:validation:Optional
	MimeType *string `json:"mimeType,omitempty" tf:"mime_type,omitempty"`
}

type TriggerObservation struct {
}

type TriggerParameters struct {

	// The absolute number of time series
	// that must fail the predicate for the
	// condition to be triggered.
	// +kubebuilder:validation:Optional
	Count *float64 `json:"count,omitempty" tf:"count,omitempty"`

	// The percentage of time series that
	// must fail the predicate for the
	// condition to be triggered.
	// +kubebuilder:validation:Optional
	Percent *float64 `json:"percent,omitempty" tf:"percent,omitempty"`
}

// AlertPolicySpec defines the desired state of AlertPolicy
type AlertPolicySpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     AlertPolicyParameters `json:"forProvider"`
}

// AlertPolicyStatus defines the observed state of AlertPolicy.
type AlertPolicyStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        AlertPolicyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// AlertPolicy is the Schema for the AlertPolicys API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,gcpjet}
type AlertPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AlertPolicySpec   `json:"spec"`
	Status            AlertPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AlertPolicyList contains a list of AlertPolicys
type AlertPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlertPolicy `json:"items"`
}

// Repository type metadata.
var (
	AlertPolicy_Kind             = "AlertPolicy"
	AlertPolicy_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: AlertPolicy_Kind}.String()
	AlertPolicy_KindAPIVersion   = AlertPolicy_Kind + "." + CRDGroupVersion.String()
	AlertPolicy_GroupVersionKind = CRDGroupVersion.WithKind(AlertPolicy_Kind)
)

func init() {
	SchemeBuilder.Register(&AlertPolicy{}, &AlertPolicyList{})
}
