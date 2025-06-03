document.getElementById('configForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    console.log('Form submit event triggered'); // Debug log

    const statusDiv = document.getElementById('status');
    statusDiv.textContent = '正在保存配置...';
    statusDiv.classList.remove('text-red-600', 'text-green-600');
    statusDiv.classList.add('text-gray-600');

    const domains = document.getElementById('domains').value.split(',').map(d => d.trim()).filter(d => d !== '');
    const config = {
        globalSettings: {
            updateIntervalSeconds: parseInt(document.getElementById('updateInterval').value) || 300,
            networkInterface: document.getElementById('networkInterface').value || 'wlan0',
            ipType: document.getElementById('ipType').value || 'ipv6'
        },
        domainList: {
            domains,
            token: document.getElementById('token').value || ''
        },
        notificationSettings: {
            telegramBotToken: document.getElementById('telegramBotToken').value || '',
            telegramChatID: document.getElementById('telegramChatID').value || ''
        }
    };

    try {
        const response = await fetch('/api/config', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(config)
        });

        if (response.ok) {
            statusDiv.textContent = '配置保存成功！IP 更新程序已启动或更新。';
            statusDiv.classList.remove('text-gray-600', 'text-red-600');
            statusDiv.classList.add('text-green-600');
            // Reload configuration to update form
            const loadResponse = await fetch('/api/config');
            if (loadResponse.ok) {
                const loadedConfig = await loadResponse.json();
                if (loadedConfig.domainList && loadedConfig.globalSettings && loadedConfig.notificationSettings) {
                    document.getElementById('domains').value = loadedConfig.domainList.domains.join(', ') || '';
                    document.getElementById('token').value = loadedConfig.domainList.token || '';
                    document.getElementById('updateInterval').value = loadedConfig.globalSettings.updateIntervalSeconds || 300;
                    document.getElementById('networkInterface').value = loadedConfig.globalSettings.networkInterface || 'wlan0';
                    document.getElementById('ipType').value = loadedConfig.globalSettings.ipType || 'ipv6';
                    document.getElementById('telegramBotToken').value = loadedConfig.notificationSettings.telegramBotToken || '';
                    document.getElementById('telegramChatID').value = loadedConfig.notificationSettings.telegramChatID || '';
                } else {
                    console.log('Received empty or incomplete config, keeping form as is');
                }
            } else {
                console.error('Failed to reload config after save');
            }
        } else {
            const error = await response.text();
            statusDiv.textContent = `保存失败：${error}`;
            statusDiv.classList.remove('text-gray-600', 'text-green-600');
            statusDiv.classList.add('text-red-600');
        }
    } catch (error) {
        statusDiv.textContent = `错误：${error.message}`;
        statusDiv.classList.remove('text-gray-600', 'text-green-600');
        statusDiv.classList.add('text-red-600');
    }
});

// Load existing configuration
window.addEventListener('load', async () => {
    try {
        const response = await fetch('/api/config');
        if (response.ok) {
            const config = await response.json();
            if (config.domainList && config.globalSettings && config.notificationSettings) {
                document.getElementById('domains').value = config.domainList.domains.join(', ') || '';
                document.getElementById('token').value = config.domainList.token || '';
                document.getElementById('updateInterval').value = config.globalSettings.updateIntervalSeconds || 300;
                document.getElementById('networkInterface').value = config.globalSettings.networkInterface || 'wlan0';
                document.getElementById('ipType').value = config.globalSettings.ipType || 'ipv6';
                document.getElementById('telegramBotToken').value = config.notificationSettings.telegramBotToken || '';
                document.getElementById('telegramChatID').value = config.notificationSettings.telegramChatID || '';
                document.getElementById('status').textContent = '已加载现有配置';
                document.getElementById('status').classList.add('text-green-600');
            } else {
                document.getElementById('status').textContent = '未找到现有配置，请输入并保存';
                document.getElementById('status').classList.add('text-gray-600');
            }
        } else {
            document.getElementById('status').textContent = '未找到现有配置，请输入并保存';
            document.getElementById('status').classList.add('text-gray-600');
        }
    } catch (error) {
        document.getElementById('status').textContent = `加载配置失败：${error.message}`;
        document.getElementById('status').classList.add('text-red-600');
    }
});
