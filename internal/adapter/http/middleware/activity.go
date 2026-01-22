package middleware

import "context"

type userActivity struct {
	UserID string
	Role   string
}

const userActivityKey contextKey = "user_activity"

func withUserActivity(ctx context.Context) (context.Context, *userActivity) {
	activity := &userActivity{}
	return context.WithValue(ctx, userActivityKey, activity), activity
}

func userActivityFromContext(ctx context.Context) (*userActivity, bool) {
	activity, ok := ctx.Value(userActivityKey).(*userActivity)
	return activity, ok
}
