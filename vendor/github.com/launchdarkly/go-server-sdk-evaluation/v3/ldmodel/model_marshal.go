package ldmodel

import (
	"github.com/launchdarkly/go-jsonstream/v3/jwriter"
	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/ldcontext"
)

// For backward compatibility, we are only allowed to drop out properties that have default values if
// Go SDK 4.x would also have done so (since some SDKs are not tolerant of missing properties in
// general). This is true of all properties that have OptionalInt-like semantics (having either a
// numeric value or "undefined"), and properties that could be either a JSON object or null (like
// VariationOrRollout.Rollout), and the BucketBy property which has optional-string-like behavior.
// Array properties should not be dropped even if nil.
//
// Properties that did not exist prior to Go SDK v5 are always safe to drop if they have default
// values, since older SDKs will never look for them. These are:
//
// - FeatureFlag.ClientSideAvailability
// - FeatureFlag.Migration
// - FeatureFlag.Migration.CheckRatio
// - FeatureFlag.SamplingRatio
// - FeatureFlag.ExcludeFromSummaries
//
// - Segment.Unbounded

func marshalFeatureFlag(flag FeatureFlag) ([]byte, error) {
	w := jwriter.NewWriter()
	marshalFeatureFlagToWriter(flag, &w)
	return w.Bytes(), w.Error()
}

func marshalFeatureFlagToWriter(flag FeatureFlag, w *jwriter.Writer) {
	obj := w.Object()

	obj.Name("key").String(flag.Key)

	obj.Name("on").Bool(flag.On)

	prereqsArr := obj.Name("prerequisites").Array()
	for _, p := range flag.Prerequisites {
		prereqObj := prereqsArr.Object()
		prereqObj.Name("key").String(p.Key)
		prereqObj.Name("variation").Int(p.Variation)
		prereqObj.End()
	}
	prereqsArr.End()

	writeTargets(&obj, flag.Targets, "targets")
	writeTargets(&obj, flag.ContextTargets, "contextTargets")

	rulesArr := obj.Name("rules").Array()
	for _, r := range flag.Rules {
		ruleObj := rulesArr.Object()
		writeVariationOrRolloutProperties(&ruleObj, r.VariationOrRollout)
		ruleObj.Maybe("id", r.ID != "").String(r.ID)
		writeClauses(w, &ruleObj, r.Clauses)
		ruleObj.Name("trackEvents").Bool(r.TrackEvents)
		ruleObj.End()
	}
	rulesArr.End()

	fallthroughObj := obj.Name("fallthrough").Object()
	writeVariationOrRolloutProperties(&fallthroughObj, flag.Fallthrough)
	fallthroughObj.End()

	flag.OffVariation.WriteToJSONWriter(obj.Name("offVariation"))

	variationsArr := obj.Name("variations").Array()
	for _, v := range flag.Variations {
		v.WriteToJSONWriter(w)
	}
	variationsArr.End()

	// In the older JSON schema, ClientSideAvailability.UsingEnvironmentID was in "clientSide", and
	// ClientSideAvailability.UsingMobileKey was assumed to be true. In the newer schema, those are
	// both in a "clientSideAvailability" object.
	//
	// If ClientSideAvailability.Explicit is true, then this flag used the newer schema and should be
	// reserialized the same way. If it is false, we will reserialize with the old schema, which
	// does not include UsingMobileKey; note that in that case UsingMobileKey is assumed to be true.
	//
	// For backward compatibility with older SDKs that might be reading a flag that was serialized by
	// this SDK, we always include the older "clientSide" property if it would be true.
	if flag.ClientSideAvailability.Explicit {
		csaObj := obj.Name("clientSideAvailability").Object()
		csaObj.Name("usingMobileKey").Bool(flag.ClientSideAvailability.UsingMobileKey)
		csaObj.Name("usingEnvironmentId").Bool(flag.ClientSideAvailability.UsingEnvironmentID)
		csaObj.End()
	}
	obj.Name("clientSide").Bool(flag.ClientSideAvailability.UsingEnvironmentID)

	obj.Name("salt").String(flag.Salt)

	obj.Name("trackEvents").Bool(flag.TrackEvents)
	obj.Name("trackEventsFallthrough").Bool(flag.TrackEventsFallthrough)

	obj.Name("debugEventsUntilDate").Float64OrNull(flag.DebugEventsUntilDate != 0, float64(flag.DebugEventsUntilDate))

	obj.Name("version").Int(flag.Version)

	obj.Name("deleted").Bool(flag.Deleted)

	if flag.Migration != nil {
		migrationObj := obj.Name("migration").Object()

		if checkRatio, ok := flag.Migration.CheckRatio.Get(); ok {
			migrationObj.Name("checkRatio").Int(checkRatio)
		}

		migrationObj.End()
	}

	if weight, ok := flag.SamplingRatio.Get(); ok {
		obj.Name("samplingRatio").Int(weight)
	}

	if flag.ExcludeFromSummaries {
		obj.Name("excludeFromSummaries").Bool(flag.ExcludeFromSummaries)
	}

	obj.End()
}

func writeTargets(obj *jwriter.ObjectState, targets []Target, name string) {
	targetsArr := obj.Name(name).Array()
	for _, t := range targets {
		targetObj := targetsArr.Object()
		if t.ContextKind != "" {
			targetObj.Name("contextKind").String(string(t.ContextKind))
		}
		targetObj.Name("variation").Int(t.Variation)
		writeStringArray(&targetObj, "values", t.Values)
		targetObj.End()
	}
	targetsArr.End()
}

func marshalSegment(segment Segment) ([]byte, error) {
	w := jwriter.NewWriter()
	marshalSegmentToWriter(segment, &w)
	return w.Bytes(), w.Error()
}

func marshalSegmentToWriter(segment Segment, w *jwriter.Writer) {
	obj := w.Object()

	obj.Name("key").String(segment.Key)
	writeStringArray(&obj, "included", segment.Included)
	writeStringArray(&obj, "excluded", segment.Excluded)
	writeSegmentTargets(&obj, segment.IncludedContexts, "includedContexts")
	writeSegmentTargets(&obj, segment.ExcludedContexts, "excludedContexts")
	obj.Name("salt").String(segment.Salt)

	rulesArr := obj.Name("rules").Array()
	for _, r := range segment.Rules {
		ruleObj := rulesArr.Object()
		ruleObj.Name("id").String(r.ID)
		writeClauses(w, &ruleObj, r.Clauses)
		ruleObj.Maybe("weight", r.Weight.IsDefined()).Int(r.Weight.IntValue())
		writeAttrRef(ruleObj.Maybe("bucketBy", r.BucketBy.IsDefined()), &r.BucketBy, r.RolloutContextKind)
		ruleObj.Maybe("rolloutContextKind", r.RolloutContextKind != "").String(string(r.RolloutContextKind))
		ruleObj.End()
	}
	rulesArr.End()

	obj.Maybe("unbounded", segment.Unbounded).Bool(segment.Unbounded)
	obj.Maybe("unboundedContextKind", segment.UnboundedContextKind != "").String(string(segment.UnboundedContextKind))

	obj.Name("version").Int(segment.Version)
	segment.Generation.WriteToJSONWriter(obj.Name("generation"))
	obj.Name("deleted").Bool(segment.Deleted)

	obj.End()
}

func writeSegmentTargets(obj *jwriter.ObjectState, targets []SegmentTarget, name string) {
	targetsArr := obj.Name(name).Array()
	for _, t := range targets {
		targetObj := targetsArr.Object()
		if t.ContextKind != "" {
			targetObj.Name("contextKind").String(string(t.ContextKind))
		}
		writeStringArray(&targetObj, "values", t.Values)
		targetObj.End()
	}
	targetsArr.End()
}

func writeStringArray(obj *jwriter.ObjectState, name string, values []string) {
	arr := obj.Name(name).Array()
	for _, v := range values {
		arr.String(v)
	}
	arr.End()
}

func writeVariationOrRolloutProperties(obj *jwriter.ObjectState, vr VariationOrRollout) {
	obj.Maybe("variation", vr.Variation.IsDefined()).Int(vr.Variation.IntValue())
	if len(vr.Rollout.Variations) > 0 {
		rolloutObj := obj.Name("rollout").Object()
		rolloutObj.Maybe("kind", vr.Rollout.Kind != "").String(string(vr.Rollout.Kind))
		rolloutObj.Maybe("contextKind", vr.Rollout.ContextKind != "").String(string(vr.Rollout.ContextKind))
		variationsArr := rolloutObj.Name("variations").Array()
		for _, wv := range vr.Rollout.Variations {
			variationObj := variationsArr.Object()
			variationObj.Name("variation").Int(wv.Variation)
			variationObj.Name("weight").Int(wv.Weight)
			variationObj.Maybe("untracked", wv.Untracked).Bool(wv.Untracked)
			variationObj.End()
		}
		variationsArr.End()
		rolloutObj.Maybe("seed", vr.Rollout.Seed.IsDefined()).Int(vr.Rollout.Seed.IntValue())
		writeAttrRef(rolloutObj.Maybe("bucketBy", vr.Rollout.BucketBy.IsDefined()),
			&vr.Rollout.BucketBy, vr.Rollout.ContextKind)
		rolloutObj.End()
	}
}

func writeClauses(w *jwriter.Writer, obj *jwriter.ObjectState, clauses []Clause) {
	clausesArr := obj.Name("clauses").Array()
	for _, c := range clauses {
		clauseObj := clausesArr.Object()
		if c.ContextKind != "" {
			clauseObj.Name("contextKind").String(string(c.ContextKind))
		}

		clauseObj.Name("attribute")
		if !c.Attribute.IsDefined() {
			// See comments in unmarshaling logic - for consistency with LD service behavior, we serialize
			// this as an empty string even if it was undefined
			w.String("")
		} else {
			writeAttrRef(w, &c.Attribute, c.ContextKind)
		}

		clauseObj.Name("op").String(string(c.Op))
		valuesArr := clauseObj.Name("values").Array()
		for _, v := range c.Values {
			v.WriteToJSONWriter(w)
		}
		valuesArr.End()
		clauseObj.Name("negate").Bool(c.Negate)
		clauseObj.End()
	}
	clausesArr.End()
}

func writeAttrRef(w *jwriter.Writer, ref *ldattr.Ref, contextKind ldcontext.Kind) {
	if contextKind == "" {
		// If there was no context kind, then we received this as old-style data in which the attribute is
		// interpreted as a plain name rather than an attribute reference. However, ref.String() will always
		// return an attribute reference which could contain escape characters. We should instead get the
		// unescaped attribute name, which AttrRef represents as the first element in a single-element path.
		w.String(ref.Component(0))
	} else {
		w.String(ref.String())
	}
}
