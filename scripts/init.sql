-- 监控交易员表
CREATE TABLE trader_monitors (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    trader_user_id VARCHAR(50) NOT NULL UNIQUE COMMENT '交易员ID',
    trader_name VARCHAR(100) COMMENT '交易员昵称',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用监控',
    monitor_interval INT DEFAULT 30 COMMENT '监控间隔(秒)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_trader_user_id (trader_user_id),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='监控交易员配置表';

-- 订单历史记录表
CREATE TABLE order_history (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    trader_user_id VARCHAR(50) NOT NULL COMMENT '交易员ID',
    order_id VARCHAR(50) NOT NULL COMMENT '订单ID',
    order_data JSON COMMENT '完整订单JSON数据',
    status ENUM('ACTIVE', 'CLOSED') DEFAULT 'ACTIVE' COMMENT '订单状态',
    position_side VARCHAR(10) COMMENT '持仓方向',
    open_size DECIMAL(20,8) COMMENT '开仓数量',
    open_price DECIMAL(20,8) COMMENT '开仓价格',
    open_leverage VARCHAR(10) COMMENT '杠杆倍数',
    first_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '首次发现时间',
    last_seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
    closed_at TIMESTAMP NULL COMMENT '平仓时间',
    UNIQUE KEY uk_trader_order (trader_user_id, order_id),
    INDEX idx_trader_user_id (trader_user_id),
    INDEX idx_status (status),
    INDEX idx_first_seen_at (first_seen_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单历史记录表';

-- 通知记录表
CREATE TABLE notification_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    trader_user_id VARCHAR(50) NOT NULL,
    order_id VARCHAR(50),
    notification_type ENUM('NEW_ORDER', 'ORDER_CLOSED') NOT NULL,
    message TEXT,
    status ENUM('PENDING', 'SUCCESS', 'FAILED') DEFAULT 'PENDING',
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    error_msg TEXT,
    INDEX idx_trader_user_id (trader_user_id),
    INDEX idx_notification_type (notification_type),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知发送记录表';
