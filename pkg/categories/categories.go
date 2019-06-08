package categories

import (
	"log"
	"time"

	u "github.com/hoenn/ynab-metrics/pkg/units"

	"github.com/prometheus/client_golang/prometheus"
	"go.bmvs.io/ynab"
	"go.bmvs.io/ynab/api/budget"
	"go.bmvs.io/ynab/api/category"
)

var categoryBudgeted = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "category_budget",
	Help: "Category budget gauge",
},
	[]string{"budget_name", "name", "group_name"})

var categoryActivity = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "category_activity",
	Help: "Category activity gauge",
},
	[]string{"budget_name", "name", "group_name"})

var categoryBalance = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "category_balance",
	Help: "Category balance gauge",
},
	[]string{"budget_name", "name", "group_name"})

var categoryGoalTarget = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "category_goal_target",
	Help: "Category goal target gauge",
},
	[]string{"budget_name", "name", "group_name"})

var categoryMonthlyGoalTarget = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "category_monthly_goal_target",
	Help: "Category monthly goal target gauge",
},
	[]string{"budget_name", "name", "group_name"})

func init() {
	prometheus.MustRegister(categoryBudgeted)
	prometheus.MustRegister(categoryActivity)
	prometheus.MustRegister(categoryBalance)
	prometheus.MustRegister(categoryGoalTarget)
	prometheus.MustRegister(categoryMonthlyGoalTarget)
}

//StartMetrics collects accounts metrics given a list of budgets
func StartMetrics(c ynab.ClientServicer, budgets []*budget.Budget) {
	log.Print("Getting Categories...")

	for _, b := range budgets {
		categoryGroupMap := map[string]string{}

		for _, v := range b.CategoryGroups {
			categoryGroupMap[v.ID] = v.Name
		}

		for _, c := range b.Categories {
			if !c.Deleted {
				budgetName := b.Name
				categoryName := c.Name
				categoryGroupName := categoryGroupMap[c.CategoryGroupID]

				goalTarget := float64(u.Dollars(*c.GoalTarget))
				monthlyGoalTarget := float64(0.0)

				if c.GoalType != nil {
					switch *c.GoalType {
						case category.GoalMonthlyFunding:
							monthlyGoalTarget = goalTarget
						case category.GoalTargetCategoryBalanceByDate:
							remaining := *c.GoalTarget - c.Balance + c.Budgeted
							remainingMonths := monthsDiff(time.Now(), c.GoalTargetMonth.Time)

							if remainingMonths >= 0 {
								monthlyGoalTarget = float64(u.Dollars(remaining / int64(remainingMonths + 1)))
							}
					}
				}

				categoryBudgeted.WithLabelValues(budgetName, categoryName, categoryGroupName).Set(float64(u.Dollars(c.Budgeted)))
				categoryActivity.WithLabelValues(budgetName, categoryName, categoryGroupName).Set(float64(u.Dollars(c.Activity)))
				categoryBalance.WithLabelValues(budgetName, categoryName, categoryGroupName).Set(float64(u.Dollars(c.Balance)))

				categoryGoalTarget.WithLabelValues(budgetName, categoryName, categoryGroupName).Set(goalTarget)
				categoryMonthlyGoalTarget.WithLabelValues(budgetName, categoryName, categoryGroupName).Set(monthlyGoalTarget)
			}
		}
	}
}

func monthsDiff(a, b time.Time) int {
	return int(b.Month()) - int(a.Month()) + (b.Year() - a.Year()) * 12
}
