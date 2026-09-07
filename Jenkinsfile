pipeline{
    agent any
    tools{
        go 'Go'
    }


    stages{
        stage('checkout Github'){
            steps{
                git branch: 'main', credentialsId: 'Jenkins-git', url: 'https://github.com/Lu2ky/04.2-chat-websocket.git'
            }
        }
        stage('Install dependencias'){
            steps{
                sh 'go mod tidy'
            }
        }
        stage('Tests Unitarios'){
            steps{
                sh 'go test -v -cover'
            }
        }
        stage('Test Coverage'){
            steps{
                sh 'go test -coverprofile=coverage.out'
                sh 'go tool cover -func=coverage.out'
            }
        }
        stage('Docker Build'){
            steps{
                script{
                    docker.build("websocket-for-chat:${BUILD_NUMBER}", ".")
                }
            }
        }
        stage('Docker Push (Opcional)'){
            when {
                branch 'main'
            }
            steps{
                sh "docker push tu-registry/websocket-for-chat:${BUILD_NUMBER}"
                sh "docker tag websocket-for-chat:${BUILD_NUMBER} tu-registry/websocket-for-chat:latest"
                sh "docker push tu-registry websocket-for-chat:latest"
            }
        }
    }
    post{
        success{
            echo 'Build complete'
        }
        failure{
            echo 'Build failed'
        }
    }
}