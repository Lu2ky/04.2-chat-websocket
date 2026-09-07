pipeline{
    agent any
    tools{
        go 'go1.26.7'

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
                    docker build -t websocket-app:${BUILD_NUMBER} .
                    docker tag websocket-app:${BUILD_NUMBER} websocket-app:latest
                }
                
            }
        }
        stage('Docker Push (Opcional)'){
            when {
                branch 'main'
            }
            steps{
                sh 'echo "Aquí iría: docker push tu-registry/websocket-app:${BUILD_NUMBER}"'
                sh 'echo "Y docker push registry/websocket-app:latest"'
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