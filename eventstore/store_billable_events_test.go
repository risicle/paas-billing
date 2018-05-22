package eventstore_test

import (
	"encoding/json"
	"time"

	"github.com/alphagov/paas-billing/eventio"
	"github.com/alphagov/paas-billing/eventstore"
	"github.com/alphagov/paas-billing/testenv"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetBillableEvents", func() {

	var (
		cfg eventstore.Config
	)

	BeforeEach(func() {
		cfg = testenv.BasicConfig
	})

	/*-----------------------------------------------------------------------------------*
	.                                                                                     .
	       00:00       01:00                                                             .
	         |           |                                                               .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [====app1===]   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	       start       stop                                                              .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	<=======================================PLAN1=======================================>.
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	*-----------------------------------------------------------------------------------*/
	It("Should return one BillingEvent for an app that was running for 1hr", func() {
		cfg.AddPlan(eventstore.PricingPlan{
			PlanGUID:  eventstore.ComputePlanGUID,
			ValidFrom: "2001-01-01",
			Name:      "PLAN1",
			Components: []eventstore.PricingPlanComponent{
				{
					Name:         "compute",
					Formula:      "ceil($time_in_seconds/3600) * 0.01",
					CurrencyCode: "GBP",
					VATCode:      "Standard",
				},
			},
		})

		db, err := testenv.Open(cfg)
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		app1EventStart := testenv.Row{
			"guid":        "ee28a570-f485-48e1-87d0-98b7b8b66dfa",
			"created_at":  "2001-01-01T00:00Z",
			"raw_message": json.RawMessage(`{"state": "STARTED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
		}
		app1EventStop := testenv.Row{
			"guid":        "8d9036c5-8367-497d-bb56-94bfcac6621a",
			"created_at":  "2001-01-01T01:00Z",
			"raw_message": json.RawMessage(`{"state": "STOPPED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
		}
		Expect(db.Insert("app_usage_events", app1EventStart, app1EventStop)).To(Succeed())

		Expect(db.Schema.Refresh()).To(Succeed())

		rows, err := db.Schema.GetBillableEventRows(eventio.EventFilter{
			RangeStart: "2001-01-01",
			RangeStop:  "2001-02-01",
		})
		Expect(err).ToNot(HaveOccurred())
		defer rows.Close()

		Expect(rows.Next()).To(BeTrue(), "expected another row")
		Expect(rows.Event()).To(Equal(&eventio.BillableEvent{
			EventGUID:     "ee28a570-f485-48e1-87d0-98b7b8b66dfa",
			EventStart:    "2001-01-01T00:00:00+00:00",
			EventStop:     "2001-01-01T01:00:00+00:00",
			ResourceGUID:  "c85e98f0-6d1b-4f45-9368-ea58263165a0",
			ResourceName:  "APP1",
			ResourceType:  "app",
			OrgGUID:       "51ba75ef-edc0-47ad-a633-a8f6e8770944",
			SpaceGUID:     "276f4886-ac40-492d-a8cd-b2646637ba76",
			PlanGUID:      "f4d4b95a-f55e-4593-8d54-3364c25798c4",
			NumberOfNodes: 1,
			MemoryInMB:    1024,
			StorageInMB:   0,
			Price: eventio.Price{
				IncVAT: "0.012",
				ExVAT:  "0.01",
				Details: []eventio.PriceComponent{
					{
						Name:         "compute",
						PlanName:     "PLAN1",
						Start:        "2001-01-01T00:00:00+00:00",
						Stop:         "2001-01-01T01:00:00+00:00",
						VatRate:      "0.2",
						VatCode:      "Standard",
						CurrencyCode: "GBP",
						CurrencyRate: "1",
						IncVAT:       "0.012",
						ExVAT:        "0.01",
					},
				},
			},
		}))

		Expect(rows.Next()).To(BeFalse(), "did not expect any more rows")
	})

	/*-----------------------------------------------------------------------------------*
	.                                                                                     .
	       00:00       01:00                                                             .
	         |           |                                                               .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [====app1===]   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [====ibm1===]   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	       start       stop                                                              .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	<=======================================PLAN1=======================================>.
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	*-----------------------------------------------------------------------------------*/
	FIt("Should return one BillingEvent for a compose instance that was running for 1hr", func() {
		cfg.AddPlan(eventstore.PricingPlan{
			PlanGUID:  eventstore.ComputePlanGUID,
			ValidFrom: "2001-01-01",
			Name:      "PLAN1",
			Components: []eventstore.PricingPlanComponent{
				{
					Name:         "compose",
					Formula:      "ceil($time_in_seconds/3600) * 0.01",
					CurrencyCode: "GBP",
					VATCode:      "Standard",
				},
			},
		})

		db, err := testenv.Open(cfg)
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		app1EventStart := testenv.Row{
			"guid":        "ee28a570-f485-48e1-87d0-98b7b8b66dfa",
			"created_at":  "2001-01-01T00:00Z",
			"raw_message": json.RawMessage(`{"state": "STARTED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "IBM1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
		}
		app1EventStop := testenv.Row{
			"guid":        "8d9036c5-8367-497d-bb56-94bfcac6621a",
			"created_at":  "2001-01-01T01:00Z",
			"raw_message": json.RawMessage(`{"state": "STOPPED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "IBM1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
		}
		service1EventStart := testenv.Row{
			"guid":        "c497eb13-f48a-4859-be53-5569f302b516",
			"kind":        "service",
			"created_at":  time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
			"raw_message": json.RawMessage(` {"id": "5aba15474a64fd00141f4263", "ip": "", "data": {"units": "3", "memory": "306 MB", "cluster": "gds-eu-west1-c00", "storage": "3 GB", "deployment": "prod-8b361bfe-c5a4-4c41-bb05-6a627bff8656"}, "event": "deployment.scale.members", "_links": {"alerts": {"href": "", "templated": false}, "backups": {"href": "", "templated": false}, "cluster": {"href": "", "templated": false}, "scalings": {"href": "", "templated": false}, "portal_users": {"href": "", "templated": false}, "compose_web_ui": {"href": "", "templated": false}}, "user_id": "", "account_id": "58d3e39c0045bb00135ee6ad", "cluster_id": "5941cf9f859d2c0015000021", "created_at": "2018-03-27T09:56:23.363Z", "user_agent": "", "deployment_id": "59de3e8cc9ecc40010324fc6"}`),
		}
		Expect(db.Insert("app_usage_events", app1EventStart, app1EventStop)).To(Succeed())
		Expect(db.Insert("compose_audit_events", service1EventStart)).To(Succeed())

		Expect(db.Schema.Refresh()).To(Succeed())

		rows, err := db.Schema.GetBillableEventRows(eventio.EventFilter{
			RangeStart: "2001-01-01",
			RangeStop:  "2001-02-01",
		})
		Expect(err).ToNot(HaveOccurred())
		defer rows.Close()

		Expect(rows.Next()).To(BeTrue(), "expected another row")
		Expect(rows.Event()).To(Equal(&eventio.BillableEvent{
			EventGUID:     "ee28a570-f485-48e1-87d0-98b7b8b66dfa",
			EventStart:    "2001-01-01T00:00:00+00:00",
			EventStop:     "2001-01-01T01:00:00+00:00",
			ResourceGUID:  "c85e98f0-6d1b-4f45-9368-ea58263165a0",
			ResourceName:  "IBM1",
			ResourceType:  "task",
			OrgGUID:       "51ba75ef-edc0-47ad-a633-a8f6e8770944",
			SpaceGUID:     "276f4886-ac40-492d-a8cd-b2646637ba76",
			PlanGUID:      "8d3383cf-9477-46cc-a219-ec0c23c020dd",
			NumberOfNodes: 1,
			MemoryInMB:    1024,
			StorageInMB:   0,
			Price: eventio.Price{
				IncVAT: "0.012",
				ExVAT:  "0.01",
				Details: []eventio.PriceComponent{
					{
						Name:         "compose",
						PlanName:     "PLAN1",
						Start:        "2001-01-01T00:00:00+00:00",
						Stop:         "2001-01-01T01:00:00+00:00",
						VatRate:      "0.2",
						VatCode:      "Standard",
						CurrencyCode: "GBP",
						CurrencyRate: "1",
						IncVAT:       "0.012",
						ExVAT:        "0.01",
					},
				},
			},
		}))

		Expect(rows.Next()).To(BeFalse(), "did not expect any more rows")
	})

	/*-----------------------------------------------------------------------------------*
	       00:00       01:00       02:00                                                 .
	         |           |           |                                                   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [=========APP1==========]   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	       start      scale+1      stop                                                  .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	<=======================================PLAN1=======================================>.
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	-------------------------------------------------------------------------------------*/
	It("Should return two BillingEvent that represent a scaling", func() {
		cfg.AddPlan(eventstore.PricingPlan{
			PlanGUID:  eventstore.ComputePlanGUID,
			ValidFrom: "2001-01-01",
			Name:      "PLAN1",
			Components: []eventstore.PricingPlanComponent{
				{
					Name:         "compute",
					Formula:      "ceil($time_in_seconds/3600) * 0.01",
					CurrencyCode: "GBP",
					VATCode:      "Standard",
				},
			},
		})

		db, err := testenv.Open(cfg)
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		app1EventStart := testenv.Row{
			"guid":        "aa30fa3c-725d-4272-9052-c7186d4968a6",
			"created_at":  "2001-01-01T00:00Z",
			"raw_message": json.RawMessage(`{"state": "STARTED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
		}
		app1EventScale := testenv.Row{
			"guid":        "be28a570-f485-48e1-87d0-98b7b8b66dfa",
			"created_at":  "2001-01-01T01:00Z",
			"raw_message": json.RawMessage(`{"state": "STARTED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 2, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
		}
		app1EventStop := testenv.Row{
			"guid":        "cd9036c5-8367-497d-bb56-94bfcac6621a",
			"created_at":  "2001-01-01T02:00Z",
			"raw_message": json.RawMessage(`{"state": "STOPPED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 2, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
		}
		Expect(db.Insert("app_usage_events", app1EventStart, app1EventScale, app1EventStop)).To(Succeed())

		Expect(db.Schema.Refresh()).To(Succeed())

		rows, err := db.Schema.GetBillableEventRows(eventio.EventFilter{
			RangeStart: "2001-01-01",
			RangeStop:  "2001-02-01",
		})
		Expect(err).ToNot(HaveOccurred())
		defer rows.Close()

		Expect(rows.Next()).To(BeTrue(), "expected another row")
		Expect(rows.Event()).To(Equal(&eventio.BillableEvent{
			EventGUID:     "aa30fa3c-725d-4272-9052-c7186d4968a6",
			EventStart:    "2001-01-01T00:00:00+00:00",
			EventStop:     "2001-01-01T01:00:00+00:00",
			ResourceGUID:  "c85e98f0-6d1b-4f45-9368-ea58263165a0",
			ResourceName:  "APP1",
			ResourceType:  "app",
			OrgGUID:       "51ba75ef-edc0-47ad-a633-a8f6e8770944",
			SpaceGUID:     "276f4886-ac40-492d-a8cd-b2646637ba76",
			PlanGUID:      "f4d4b95a-f55e-4593-8d54-3364c25798c4",
			NumberOfNodes: 1,
			MemoryInMB:    1024,
			StorageInMB:   0,
			Price: eventio.Price{
				IncVAT: "0.012",
				ExVAT:  "0.01",
				Details: []eventio.PriceComponent{
					{
						Name:         "compute",
						PlanName:     "PLAN1",
						Start:        "2001-01-01T00:00:00+00:00",
						Stop:         "2001-01-01T01:00:00+00:00",
						VatRate:      "0.2",
						VatCode:      "Standard",
						CurrencyCode: "GBP",
						CurrencyRate: "1",
						IncVAT:       "0.012",
						ExVAT:        "0.01",
					},
				},
			},
		}), "expected a 1hr BillingEvent before scaling")

		Expect(rows.Next()).To(BeTrue(), "expected another row")
		Expect(rows.Event()).To(Equal(&eventio.BillableEvent{
			EventGUID:     "be28a570-f485-48e1-87d0-98b7b8b66dfa",
			EventStart:    "2001-01-01T01:00:00+00:00",
			EventStop:     "2001-01-01T02:00:00+00:00",
			ResourceGUID:  "c85e98f0-6d1b-4f45-9368-ea58263165a0",
			ResourceName:  "APP1",
			ResourceType:  "app",
			OrgGUID:       "51ba75ef-edc0-47ad-a633-a8f6e8770944",
			SpaceGUID:     "276f4886-ac40-492d-a8cd-b2646637ba76",
			PlanGUID:      "f4d4b95a-f55e-4593-8d54-3364c25798c4",
			NumberOfNodes: 2,
			MemoryInMB:    1024,
			StorageInMB:   0,
			Price: eventio.Price{
				IncVAT: "0.012",
				ExVAT:  "0.01",
				Details: []eventio.PriceComponent{
					{
						Name:         "compute",
						PlanName:     "PLAN1",
						Start:        "2001-01-01T01:00:00+00:00",
						Stop:         "2001-01-01T02:00:00+00:00",
						VatRate:      "0.2",
						VatCode:      "Standard",
						CurrencyCode: "GBP",
						CurrencyRate: "1",
						IncVAT:       "0.012",
						ExVAT:        "0.01",
					},
				},
			},
		}), "expected a 1hr BillingEvent after scaling")

	})

	/*-----------------------------------------------------------------------------------*
	       00:00                                                                    now  .
	         |                                                                       |   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [===============================app1=====================================   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	       start                                                                         .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	<=======================================PLAN1=======================================>.
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	-------------------------------------------------------------------------------------*/
	It("should return a BillableEvent for an app without a stop event yet", func() {
		cfg.AddPlan(eventstore.PricingPlan{
			PlanGUID:  eventstore.ComputePlanGUID,
			ValidFrom: "2001-01-01",
			Name:      "PLAN1",
			Components: []eventstore.PricingPlanComponent{
				{
					Name:         "compute",
					Formula:      "ceil($time_in_seconds/3600) * 0.01",
					CurrencyCode: "GBP",
					VATCode:      "Standard",
				},
			},
		})

		db, err := testenv.Open(cfg)
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		Expect(db.Insert("app_usage_events",
			testenv.Row{
				"guid":        "ee28a570-f485-48e1-87d0-98b7b8b66dfa",
				"created_at":  "2001-01-01T00:00Z",
				"raw_message": json.RawMessage(`{"state": "STARTED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
			},
		)).To(Succeed())

		Expect(db.Schema.Refresh()).To(Succeed())

		events, err := db.Schema.GetBillableEvents(eventio.EventFilter{
			RangeStart: "2001-01-01",
			RangeStop:  "3000-02-01",
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(len(events)).To(BeNumerically("==", 1), "expected a single event to be returned")

		Expect(testenv.Time(events[0].EventStart)).To(
			BeTemporally("==", testenv.Time("2001-01-01T00:00:00+00:00")),
			"start time should be 00:00",
		)
		Expect(testenv.Time(events[0].EventStop)).To(
			BeTemporally("~", time.Now(), 1*time.Minute),
			"stop time should be roughly now()",
		)
	})

	/*-----------------------------------------------------------------------------------*
	     2001-01-01      2002-01-01                          2002-02-02                  .
	       00:00           01:00                              02:00                 now  .
	         |               |                                   |                   |   .
	 .   .   .   .   .   .   |   .   .   .   .   .   .   .   .   |   .   .   .   .   .   .
	 .   .   [===============================APP1=====================================   .
	 .   .   .   .   .   .   |   .   .   .   .   .   .   .   .   |   .   .   .   .   .   .
	       start             |                                   |                       .
	 .   .   .   .   .   .   |   .   .   .   .   .   .   .   .   |   .   .   .   .   .   .
	<=======================================PLAN1=======================================>.
	 .   .   .   .   .   .   |   .   .   .   .   .   .   .   .   |   .   .   .   .   .   .
	 .   .   .   .   .   .   |   .   .   .   .   .   .   .   .   |   .   .   .   .   .   .
	 .   .   .   .   .       |__________requested range__________|   .   .   .   .   .   .
	-------------------------------------------------------------------------------------*/
	It("should return a BillableEvent with duration cropped to the requeted range", func() {
		cfg.AddPlan(eventstore.PricingPlan{
			PlanGUID:  eventstore.ComputePlanGUID,
			ValidFrom: "2001-01-01",
			Name:      "PLAN1",
			Components: []eventstore.PricingPlanComponent{
				{
					Name:         "compute",
					Formula:      "ceil($time_in_seconds/3600) * 0.01",
					CurrencyCode: "GBP",
					VATCode:      "Standard",
				},
			},
		})

		db, err := testenv.Open(cfg)
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		Expect(db.Insert("app_usage_events",
			testenv.Row{
				"guid":        "ee28a570-f485-48e1-87d0-98b7b8b66dfa",
				"created_at":  "2001-01-01T00:00Z",
				"raw_message": json.RawMessage(`{"state": "STARTED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
			},
		)).To(Succeed())

		Expect(db.Schema.Refresh()).To(Succeed())

		filter := eventio.EventFilter{
			RangeStart: "2002-01-01T01:00:00Z",
			RangeStop:  "2002-02-02T02:00:00Z",
		}
		events, err := db.Schema.GetBillableEvents(filter)
		Expect(err).ToNot(HaveOccurred())

		Expect(len(events)).To(BeNumerically("==", 1), "expected a single event to be returned")

		Expect(testenv.Time(events[0].EventStart)).To(
			BeTemporally("==", testenv.Time(filter.RangeStart)),
			"start time should be same as RangeStart",
		)
		Expect(testenv.Time(events[0].EventStop)).To(
			BeTemporally("==", testenv.Time(filter.RangeStop)),
			"stop time should be same as RangeStop",
		)
	})

	/*---------------------------------------------------------------------------------------*
	     2017-01-01                        2017-02-01                           2017-03-01   .
	         |                                 |                                     |   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [===============================APP1====================================]   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [============PLAN1===============][================PLAN2================]   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	----------------------------------------------------------------------------------------*/
	It("should return one BillingEvent with two pricing components when intersects two plans", func() {
		cfg.AddPlan(eventstore.PricingPlan{
			PlanGUID:  eventstore.ComputePlanGUID,
			ValidFrom: "2017-01-01",
			Name:      "PLAN1",
			Components: []eventstore.PricingPlanComponent{
				{
					Name:         "compute",
					Formula:      "1",
					CurrencyCode: "GBP",
					VATCode:      "Standard",
				},
			},
		})
		cfg.AddPlan(eventstore.PricingPlan{
			PlanGUID:  eventstore.ComputePlanGUID,
			ValidFrom: "2017-02-01",
			Name:      "PLAN2",
			Components: []eventstore.PricingPlanComponent{
				{
					Name:         "compute",
					Formula:      "33",
					CurrencyCode: "GBP",
					VATCode:      "Standard",
				},
			},
		})

		db, err := testenv.Open(cfg)
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		Expect(db.Insert("app_usage_events",
			testenv.Row{
				"guid":        "ee28a570-f485-48e1-87d0-98b7b8b66dfa",
				"created_at":  "2017-01-01T00:00Z",
				"raw_message": json.RawMessage(`{"state": "STARTED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
			},
			testenv.Row{
				"guid":        "33d0aaad-e064-4dc7-8709-0212d96c7c3f",
				"created_at":  "2017-03-01T00:00Z",
				"raw_message": json.RawMessage(`{"state": "STOPPED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
			},
		)).To(Succeed())

		Expect(db.Schema.Refresh()).To(Succeed())

		events, err := db.Schema.GetBillableEvents(eventio.EventFilter{
			RangeStart: "2017-01-01",
			RangeStop:  "2017-03-01",
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(len(events)).To(BeNumerically("==", 1), "expected two events to be returned")

		Expect(testenv.Time(events[0].EventStart)).To(
			BeTemporally("==", testenv.Time("2017-01-01T00:00:00+00:00")),
			"start time should be 2017-01-01",
		)
		Expect(testenv.Time(events[0].EventStop)).To(
			BeTemporally("==", testenv.Time("2017-03-01T00:00:00+00:00")),
			"stop time should be 2017-03-01",
		)

		Expect(events[0].Price).To(Equal(eventio.Price{
			IncVAT: "40.8",
			ExVAT:  "34",
			Details: []eventio.PriceComponent{
				{
					Name:         "compute",
					PlanName:     "PLAN1",
					Start:        "2017-01-01T00:00:00+00:00",
					Stop:         "2017-02-01T00:00:00+00:00",
					VatRate:      "0.2",
					VatCode:      "Standard",
					CurrencyCode: "GBP",
					CurrencyRate: "1",
					IncVAT:       "1.2",
					ExVAT:        "1",
				},
				{
					Name:         "compute",
					PlanName:     "PLAN2",
					Start:        "2017-02-01T00:00:00+00:00",
					Stop:         "2017-03-01T00:00:00+00:00",
					VatRate:      "0.2",
					VatCode:      "Standard",
					CurrencyCode: "GBP",
					CurrencyRate: "1",
					IncVAT:       "39.6",
					ExVAT:        "33",
				},
			},
		}))
	})

	/*---------------------------------------------------------------------------------------*
	     2017-01-01                        2017-02-01                           2017-03-01   .
	         |                                 |                                     |   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [===============================APP1====================================]   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [==============================PLAN1====================================]   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [============VATRate1============][=============VATRate2================]   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	----------------------------------------------------------------------------------------*/
	It("should return a single BillingEvent with two pricing components when a single event intersects two VAT rates", func() {
		cfg.AddPlan(eventstore.PricingPlan{
			PlanGUID:  eventstore.ComputePlanGUID,
			ValidFrom: "2017-01-01",
			Name:      "PLAN1",
			Components: []eventstore.PricingPlanComponent{
				{
					Name:         "compute",
					Formula:      "1",
					CurrencyCode: "GBP",
					VATCode:      "Standard",
				},
			},
		})
		cfg.AddVATRate(eventstore.VATRate{
			Code:      "Standard",
			Rate:      0,
			ValidFrom: "2017-02-01",
		})

		db, err := testenv.Open(cfg)
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		Expect(db.Insert("app_usage_events",
			testenv.Row{
				"guid":        "ee28a570-f485-48e1-87d0-98b7b8b66dfa",
				"created_at":  "2017-01-01T00:00Z",
				"raw_message": json.RawMessage(`{"state": "STARTED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
			},
			testenv.Row{
				"guid":        "33d0aaad-e064-4dc7-8709-0212d96c7c3f",
				"created_at":  "2017-03-01T00:00Z",
				"raw_message": json.RawMessage(`{"state": "STOPPED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
			},
		)).To(Succeed())

		Expect(db.Schema.Refresh()).To(Succeed())

		events, err := db.Schema.GetBillableEvents(eventio.EventFilter{
			RangeStart: "2017-01-01",
			RangeStop:  "2017-03-01",
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(len(events)).To(BeNumerically("==", 1), "expected a single event to be returned")

		Expect(testenv.Time(events[0].EventStart)).To(
			BeTemporally("==", testenv.Time("2017-01-01T00:00:00+00:00")),
			"start time should be 2017-01-01",
		)
		Expect(testenv.Time(events[0].EventStop)).To(
			BeTemporally("==", testenv.Time("2017-03-01T00:00:00+00:00")),
			"stop time should be 2017-03-01",
		)

		Expect(events[0].Price).To(Equal(eventio.Price{
			IncVAT: "2.2",
			ExVAT:  "2",
			Details: []eventio.PriceComponent{
				{
					Name:         "compute",
					PlanName:     "PLAN1",
					Start:        "2017-01-01T00:00:00+00:00",
					Stop:         "2017-02-01T00:00:00+00:00",
					VatRate:      "0.2",
					VatCode:      "Standard",
					CurrencyCode: "GBP",
					CurrencyRate: "1",
					IncVAT:       "1.2",
					ExVAT:        "1",
				},
				{
					Name:         "compute",
					PlanName:     "PLAN1",
					Start:        "2017-02-01T00:00:00+00:00",
					Stop:         "2017-03-01T00:00:00+00:00",
					VatRate:      "0",
					VatCode:      "Standard",
					CurrencyCode: "GBP",
					CurrencyRate: "1",
					IncVAT:       "1",
					ExVAT:        "1",
				},
			},
		}))
	})

	/*---------------------------------------------------------------------------------------*
	     2017-01-01                        2017-02-01                           2017-03-01   .
	         |                                 |                                     |   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [===============================APP1====================================]   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [==============================PLAN1====================================]   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [==========CurrencyRate1==========][==========CurrencyRate2=============]   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	----------------------------------------------------------------------------------------*/
	It("should return a single BillingEvent with two pricing components when the event intersects two CurrencyRates", func() {
		cfg.AddPlan(eventstore.PricingPlan{
			PlanGUID:  eventstore.ComputePlanGUID,
			ValidFrom: "2017-01-01",
			Name:      "PLAN1",
			Components: []eventstore.PricingPlanComponent{
				{
					Name:         "compute",
					Formula:      "1",
					CurrencyCode: "GBP",
					VATCode:      "Standard",
				},
			},
		})
		cfg.AddCurrencyRate(eventstore.CurrencyRate{
			Code:      "GBP",
			Rate:      2,
			ValidFrom: "2017-02-01",
		})

		db, err := testenv.Open(cfg)
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		Expect(db.Insert("app_usage_events",
			testenv.Row{
				"guid":        "ee28a570-f485-48e1-87d0-98b7b8b66dfa",
				"created_at":  "2017-01-01T00:00Z",
				"raw_message": json.RawMessage(`{"state": "STARTED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
			},
			testenv.Row{
				"guid":        "33d0aaad-e064-4dc7-8709-0212d96c7c3f",
				"created_at":  "2017-03-01T00:00Z",
				"raw_message": json.RawMessage(`{"state": "STOPPED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
			},
		)).To(Succeed())

		Expect(db.Schema.Refresh()).To(Succeed())

		events, err := db.Schema.GetBillableEvents(eventio.EventFilter{
			RangeStart: "2017-01-01",
			RangeStop:  "2017-03-01",
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(len(events)).To(BeNumerically("==", 1), "expected a single event to be returned")

		Expect(testenv.Time(events[0].EventStart)).To(
			BeTemporally("==", testenv.Time("2017-01-01T00:00:00+00:00")),
			"start time should be 2017-01-01",
		)
		Expect(testenv.Time(events[0].EventStop)).To(
			BeTemporally("==", testenv.Time("2017-03-01T00:00:00+00:00")),
			"stop time should be 2017-03-01",
		)

		Expect(events[0].Price).To(Equal(eventio.Price{
			IncVAT: "3.6",
			ExVAT:  "3",
			Details: []eventio.PriceComponent{
				{
					Name:         "compute",
					PlanName:     "PLAN1",
					Start:        "2017-01-01T00:00:00+00:00",
					Stop:         "2017-02-01T00:00:00+00:00",
					VatRate:      "0.2",
					VatCode:      "Standard",
					CurrencyCode: "GBP",
					CurrencyRate: "1",
					IncVAT:       "1.2",
					ExVAT:        "1",
				},
				{
					Name:         "compute",
					PlanName:     "PLAN1",
					Start:        "2017-02-01T00:00:00+00:00",
					Stop:         "2017-03-01T00:00:00+00:00",
					VatRate:      "0.2",
					VatCode:      "Standard",
					CurrencyCode: "GBP",
					CurrencyRate: "2",
					IncVAT:       "2.4",
					ExVAT:        "2",
				},
			},
		}))
	})

	/*---------------------------------------------------------------------------------------*
	     2017-01-01           2017-02-01   2017-03-01     2017-04-01            2017-05-01   .
	         |                     |           |              |                      |   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [===============================APP1====================================]   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [==============================PLAN1====================================]   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [=============VATRate1============][=============VatRate2===============]   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [====CurrencyRate1====][=====CurrencyRate2=======][====CurrencyRate3 ===]   .   .
	 .   .   .   .   .   .   .   .   .   .   .    .   .   .   .   .   .   .   .   .      .   .
	 .   .   |   .   .   .   .   . | .   .   .  | .   .   .    |  .   .   .   .   .  |   .   .
	 .   .   +-----------------------------------------------------------------------    .   .
	 .   .   |   .  component1.  . | component2 |  component3  |    component4    .  |   .   .
	----------------------------------------------------------------------------------------*/
	It("should return a single BillingEvent with four pricing components when the events intersects currency and vate rate changes", func() {
		cfg.AddPlan(eventstore.PricingPlan{
			PlanGUID:  eventstore.ComputePlanGUID,
			ValidFrom: "2017-01-01",
			Name:      "PLAN1",
			Components: []eventstore.PricingPlanComponent{
				{
					Name:         "compute",
					Formula:      "1",
					CurrencyCode: "GBP",
					VATCode:      "Standard",
				},
			},
		})
		cfg.AddVATRate(eventstore.VATRate{
			Code:      "Standard",
			Rate:      0,
			ValidFrom: "2017-03-01",
		})
		cfg.AddCurrencyRate(eventstore.CurrencyRate{
			Code:      "GBP",
			Rate:      2,
			ValidFrom: "2017-02-01",
		})
		cfg.AddCurrencyRate(eventstore.CurrencyRate{
			Code:      "GBP",
			Rate:      4,
			ValidFrom: "2017-04-01",
		})

		db, err := testenv.Open(cfg)
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		Expect(db.Insert("app_usage_events",
			testenv.Row{
				"guid":        "ee28a570-f485-48e1-87d0-98b7b8b66dfa",
				"created_at":  "2017-01-01T00:00Z",
				"raw_message": json.RawMessage(`{"state": "STARTED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
			},
			testenv.Row{
				"guid":        "33d0aaad-e064-4dc7-8709-0212d96c7c3f",
				"created_at":  "2017-05-01T00:00Z",
				"raw_message": json.RawMessage(`{"state": "STOPPED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
			},
		)).To(Succeed())

		Expect(db.Schema.Refresh()).To(Succeed())

		events, err := db.Schema.GetBillableEvents(eventio.EventFilter{
			RangeStart: "2017-01-01",
			RangeStop:  "2017-05-01",
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(len(events)).To(BeNumerically("==", 1), "expected a single event to be returned")

		Expect(testenv.Time(events[0].EventStart)).To(
			BeTemporally("==", testenv.Time("2017-01-01T00:00:00+00:00")),
			"start time should be 2017-01-01",
		)
		Expect(testenv.Time(events[0].EventStop)).To(
			BeTemporally("==", testenv.Time("2017-05-01T00:00:00+00:00")),
			"stop time should be 2017-05-01",
		)

		Expect(events[0].Price).To(Equal(eventio.Price{
			IncVAT: "9.6",
			ExVAT:  "9",
			Details: []eventio.PriceComponent{
				{
					Name:         "compute",
					PlanName:     "PLAN1",
					Start:        "2017-01-01T00:00:00+00:00",
					Stop:         "2017-02-01T00:00:00+00:00",
					VatRate:      "0.2",
					VatCode:      "Standard",
					CurrencyCode: "GBP",
					CurrencyRate: "1",
					IncVAT:       "1.2",
					ExVAT:        "1",
				},
				{
					Name:         "compute",
					PlanName:     "PLAN1",
					Start:        "2017-02-01T00:00:00+00:00",
					Stop:         "2017-03-01T00:00:00+00:00",
					VatRate:      "0.2",
					VatCode:      "Standard",
					CurrencyCode: "GBP",
					CurrencyRate: "2",
					IncVAT:       "2.4",
					ExVAT:        "2",
				},
				{
					Name:         "compute",
					PlanName:     "PLAN1",
					Start:        "2017-03-01T00:00:00+00:00",
					Stop:         "2017-04-01T00:00:00+00:00",
					VatRate:      "0",
					VatCode:      "Standard",
					CurrencyCode: "GBP",
					CurrencyRate: "2",
					IncVAT:       "2",
					ExVAT:        "2",
				},
				{
					Name:         "compute",
					PlanName:     "PLAN1",
					Start:        "2017-04-01T00:00:00+00:00",
					Stop:         "2017-05-01T00:00:00+00:00",
					VatRate:      "0",
					VatCode:      "Standard",
					CurrencyCode: "GBP",
					CurrencyRate: "4",
					IncVAT:       "4",
					ExVAT:        "4",
				},
			},
		}))
	})

	/*-----------------------------------------------------------------------------------*
	 .                                                                                   .
	  00:00:00  00:00:01                                                                 .
	       |      |                                                                      .
	 .   .  \    /   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   [==]    .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	     start  stop                                                                     .
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	<=======================================PLAN1=======================================>.
	 .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .   .
	*-----------------------------------------------------------------------------------*/
	It("Should round price of BillingEvent to a 1p if it is less than that", func() {
		cfg.AddPlan(eventstore.PricingPlan{
			PlanGUID:  eventstore.ComputePlanGUID,
			ValidFrom: "2001-01-01",
			Name:      "PLAN1",
			Components: []eventstore.PricingPlanComponent{
				{
					Name:         "compute",
					Formula:      "0.0000000001",
					CurrencyCode: "GBP",
					VATCode:      "Standard",
				},
			},
		})

		db, err := testenv.Open(cfg)
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		app1EventStart := testenv.Row{
			"guid":        "ee28a570-f485-48e1-87d0-98b7b8b66dfa",
			"created_at":  "2001-01-01T00:00:00Z",
			"raw_message": json.RawMessage(`{"state": "STARTED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
		}
		app1EventStop := testenv.Row{
			"guid":        "8d9036c5-8367-497d-bb56-94bfcac6621a",
			"created_at":  "2001-01-01T00:00:01Z",
			"raw_message": json.RawMessage(`{"state": "STOPPED", "app_guid": "c85e98f0-6d1b-4f45-9368-ea58263165a0", "app_name": "APP1", "org_guid": "51ba75ef-edc0-47ad-a633-a8f6e8770944", "space_guid": "276f4886-ac40-492d-a8cd-b2646637ba76", "space_name": "ORG1-SPACE1", "process_type": "web", "instance_count": 1, "previous_state": "STARTED", "memory_in_mb_per_instance": 1024}`),
		}
		Expect(db.Insert("app_usage_events", app1EventStart, app1EventStop)).To(Succeed())

		Expect(db.Schema.Refresh()).To(Succeed())

		events, err := db.Schema.GetBillableEvents(eventio.EventFilter{
			RangeStart: "2001-01-01",
			RangeStop:  "2001-02-01",
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(len(events)).To(BeNumerically("==", 1), "expected a single event to be returned")
		event := events[0]

		Expect(testenv.Time(event.EventStart)).To(
			BeTemporally("==", testenv.Time("2001-01-01T00:00:00+00:00")),
			"start time should be 2001-01-01 00:00:00",
		)
		Expect(testenv.Time(event.EventStop)).To(
			BeTemporally("==", testenv.Time("2001-01-01T00:00:01+00:00")),
			"stop time should be 2001-01-01 00:00:01",
		)

		Expect(event.Price).To(Equal(eventio.Price{
			IncVAT: "0.012",
			ExVAT:  "0.01",
			Details: []eventio.PriceComponent{
				{
					Name:         "compute",
					PlanName:     "PLAN1",
					Start:        "2001-01-01T00:00:00+00:00",
					Stop:         "2001-01-01T00:00:01+00:00",
					VatRate:      "0.2",
					VatCode:      "Standard",
					CurrencyCode: "GBP",
					CurrencyRate: "1",
					IncVAT:       "0.012",
					ExVAT:        "0.01",
				},
			},
		}))
	})
})