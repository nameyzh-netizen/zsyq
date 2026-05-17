package routes

import (
	"time"

	"github.com/nameyzh-netizen/zsyq/internal/handler"
	"github.com/nameyzh-netizen/zsyq/internal/handler/admin"
	"github.com/nameyzh-netizen/zsyq/internal/middleware"
	servermiddleware "github.com/nameyzh-netizen/zsyq/internal/server/middleware"
	"github.com/nameyzh-netizen/zsyq/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RegisterPaymentRoutes registers all payment-related routes:
// user-facing endpoints, webhook endpoints, and admin endpoints.
func RegisterPaymentRoutes(
	v1 *gin.RouterGroup,
	paymentHandler *handler.PaymentHandler,
	webhookHandler *handler.PaymentWebhookHandler,
	adminPaymentHandler *admin.PaymentHandler,
	jwtAuth servermiddleware.JWTAuthMiddleware,
	adminAuth servermiddleware.AdminAuthMiddleware,
	settingService *service.SettingService,
	redisClient *redis.Client,
) {
	rateLimiter := middleware.NewRateLimiter(redisClient)

	// --- User-facing payment endpoints (authenticated) ---
	authenticated := v1.Group("/payment")
	authenticated.Use(gin.HandlerFunc(jwtAuth))
	authenticated.Use(servermiddleware.BackendModeUserGuard(settingService))
	{
		authenticated.GET("/config", paymentHandler.GetPaymentConfig)
		authenticated.GET("/checkout-info", paymentHandler.GetCheckoutInfo)
		authenticated.GET("/plans", paymentHandler.GetPlans)
		authenticated.GET("/channels", paymentHandler.GetChannels)
		authenticated.GET("/limits", paymentHandler.GetLimits)

		orders := authenticated.Group("/orders")
		{
			orders.POST("", paymentHandler.CreateOrder)
			orders.POST("/verify", paymentHandler.VerifyOrder)
			orders.GET("/my", paymentHandler.GetMyOrders)
			orders.GET("/:id", paymentHandler.GetOrder)
			orders.POST("/:id/cancel", paymentHandler.CancelOrder)
			orders.POST("/:id/refund-request", paymentHandler.RequestRefund)
			orders.GET("/refund-eligible-providers", paymentHandler.GetRefundEligibleProviders)
		}
	}

	// --- Public payment endpoints (no auth, rate-limited) ---
	public := v1.Group("/payment/public")
	{
		public.POST("/orders/verify", rateLimiter.LimitWithOptions("payment-public-verify", 30, time.Minute, middleware.RateLimitOptions{
			FailureMode: middleware.RateLimitFailClose,
		}), paymentHandler.VerifyOrderPublic)
		public.POST("/orders/resolve", rateLimiter.LimitWithOptions("payment-public-resolve", 30, time.Minute, middleware.RateLimitOptions{
			FailureMode: middleware.RateLimitFailClose,
		}), paymentHandler.ResolveOrderPublicByResumeToken)
	}

	// --- Webhook endpoints (no auth, rate-limited) ---
	// Webhooks use FailOpen: if Redis is down, let the request through.
	// Blocking webhooks during Redis outages causes payment providers to
	// exhaust retries and permanently lose payment confirmations.
	webhook := v1.Group("/payment/webhook")
	{
		// EasyPay sends GET callbacks with query params
		webhook.GET("/easypay", rateLimiter.LimitWithOptions("payment-webhook-easypay", 60, time.Minute, middleware.RateLimitOptions{
			FailureMode: middleware.RateLimitFailOpen,
		}), webhookHandler.EasyPayNotify)
		webhook.POST("/easypay", rateLimiter.LimitWithOptions("payment-webhook-easypay", 60, time.Minute, middleware.RateLimitOptions{
			FailureMode: middleware.RateLimitFailOpen,
		}), webhookHandler.EasyPayNotify)
		webhook.POST("/alipay", rateLimiter.LimitWithOptions("payment-webhook-alipay", 60, time.Minute, middleware.RateLimitOptions{
			FailureMode: middleware.RateLimitFailOpen,
		}), webhookHandler.AlipayNotify)
		webhook.POST("/wxpay", rateLimiter.LimitWithOptions("payment-webhook-wxpay", 60, time.Minute, middleware.RateLimitOptions{
			FailureMode: middleware.RateLimitFailOpen,
		}), webhookHandler.WxpayNotify)
		webhook.POST("/stripe", rateLimiter.LimitWithOptions("payment-webhook-stripe", 60, time.Minute, middleware.RateLimitOptions{
			FailureMode: middleware.RateLimitFailOpen,
		}), webhookHandler.StripeWebhook)
		webhook.POST("/airwallex", rateLimiter.LimitWithOptions("payment-webhook-airwallex", 60, time.Minute, middleware.RateLimitOptions{
			FailureMode: middleware.RateLimitFailOpen,
		}), webhookHandler.AirwallexWebhook)
	}

	// --- Admin payment endpoints (admin auth) ---
	adminGroup := v1.Group("/admin/payment")
	adminGroup.Use(gin.HandlerFunc(adminAuth))
	{
		// Dashboard
		adminGroup.GET("/dashboard", adminPaymentHandler.GetDashboard)

		// Config
		adminGroup.GET("/config", adminPaymentHandler.GetConfig)
		adminGroup.PUT("/config", adminPaymentHandler.UpdateConfig)

		// Orders
		adminOrders := adminGroup.Group("/orders")
		{
			adminOrders.GET("", adminPaymentHandler.ListOrders)
			adminOrders.GET("/:id", adminPaymentHandler.GetOrderDetail)
			adminOrders.POST("/:id/cancel", adminPaymentHandler.CancelOrder)
			adminOrders.POST("/:id/retry", adminPaymentHandler.RetryFulfillment)
			adminOrders.POST("/:id/refund", adminPaymentHandler.ProcessRefund)
		}

		// Subscription Plans
		plans := adminGroup.Group("/plans")
		{
			plans.GET("", adminPaymentHandler.ListPlans)
			plans.POST("", adminPaymentHandler.CreatePlan)
			plans.PUT("/:id", adminPaymentHandler.UpdatePlan)
			plans.DELETE("/:id", adminPaymentHandler.DeletePlan)
		}

		// Provider Instances
		providers := adminGroup.Group("/providers")
		{
			providers.GET("", adminPaymentHandler.ListProviders)
			providers.POST("", adminPaymentHandler.CreateProvider)
			providers.PUT("/:id", adminPaymentHandler.UpdateProvider)
			providers.DELETE("/:id", adminPaymentHandler.DeleteProvider)
		}
	}
}