package postgres

const (
	getUserIDSQL              = `SELECT id FROM users WHERE login=$1 AND password_hash=$2`
	createUserSQL             = `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`
	createOrderSQL            = `INSERT INTO orders (user_id, number, status) VALUES ($1, $2, $3)`
	updateOrderWithAccrualSQL = `UPDATE orders SET status=$1, accrual=$2, updated_at=$3 WHERE number=$4`
	updateOrderSQL            = `UPDATE orders SET status=$1, updated_at=$2 WHERE number=$3`
	getOrderSQL               = `SELECT number, status, user_id, accrual, updated_at FROM orders WHERE number=$1`
	getAllOrdersByUserIDSQL   = `SELECT number, status, user_id, accrual, updated_at 
					   		   FROM orders 
					   		   WHERE user_id=$1 AND withdraw IS NULL ORDER BY updated_at`
	getAllOrdersByStatusSQL = `SELECT number, status, user_id, accrual, updated_at 
							   FROM orders 
							   WHERE status=$1 AND withdraw IS NULL ORDER BY updated_at`
	getBalanceSQL = `SELECT SUM(accrual) as total, SUM(withdraw) as withdraw 
                  	 FROM ( 
            			SELECT COALESCE(accrual, 0) as accrual, COALESCE(withdraw, 0) as withdraw 
            			FROM orders 
            			WHERE user_id=$1
            		 ) as t`
	withdrawBonusesSQL   = `INSERT INTO orders (user_id, number, status, withdraw) VALUES ($1, $2, $3, $4)`
	getAllWithdrawalsSQL = `SELECT number, withdraw, updated_at
			   				FROM orders 
			   				WHERE withdraw IS NOT NULL AND user_id=$1 ORDER BY updated_at`
)
