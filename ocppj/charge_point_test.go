package ocppj_test

import (
	"errors"
	"fmt"
	"github.com/lorenzodonini/ocpp-go/ocpp"
	"github.com/lorenzodonini/ocpp-go/ocppj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"strconv"
	"sync"
	"time"
)

// ----------------- Start tests -----------------

func (suite *OcppJTestSuite) TestChargePointStart() {
	suite.mockClient.On("Start", mock.AnythingOfType("string")).Return(nil)
	err := suite.chargePoint.Start("someUrl")
	assert.Nil(suite.T(), err)
}

func (suite *OcppJTestSuite) TestChargePointStartFailed() {
	suite.mockClient.On("Start", mock.AnythingOfType("string")).Return(errors.New("startError"))
	err := suite.chargePoint.Start("someUrl")
	assert.NotNil(suite.T(), err)
}

func (suite *OcppJTestSuite) TestNotStartedError() {
	t := suite.T()
	// Start normally
	req := newMockRequest("somevalue")
	err := suite.chargePoint.SendRequest(req)
	require.NotNil(t, err)
	assert.Equal(t, "ocppj client is not started, couldn't send request", err.Error())
	require.True(t, suite.clientRequestQueue.IsEmpty())
}

// ----------------- SendRequest tests -----------------

func (suite *OcppJTestSuite) TestChargePointSendRequest() {
	suite.mockClient.On("Write", mock.Anything).Return(nil)
	suite.mockClient.On("Start", mock.AnythingOfType("string")).Return(nil)
	_ = suite.chargePoint.Start("someUrl")
	mockRequest := newMockRequest("mockValue")
	err := suite.chargePoint.SendRequest(mockRequest)
	assert.Nil(suite.T(), err)
}

func (suite *OcppJTestSuite) TestChargePointSendInvalidRequest() {
	suite.mockClient.On("Write", mock.Anything).Return(nil)
	suite.mockClient.On("Start", mock.AnythingOfType("string")).Return(nil)
	_ = suite.chargePoint.Start("someUrl")
	mockRequest := newMockRequest("")
	err := suite.chargePoint.SendRequest(mockRequest)
	assert.NotNil(suite.T(), err)
}

func (suite *OcppJTestSuite) TestChargePointSendRequestFailed() {
	t := suite.T()
	var callID string
	suite.mockClient.On("Start", mock.AnythingOfType("string")).Return(nil)
	suite.mockClient.On("Write", mock.Anything).Return(errors.New("networkError")).Run(func(args mock.Arguments) {
		require.False(t, suite.clientRequestQueue.IsEmpty())
		req := suite.clientRequestQueue.Peek().(ocppj.RequestBundle)
		callID = req.Call.GetUniqueId()
		_, ok := suite.chargePoint.GetPendingRequest(callID)
		// Before anything is returned, the request must still be pending
		assert.True(t, ok)
	})
	_ = suite.chargePoint.Start("someUrl")
	mockRequest := newMockRequest("mockValue")
	err := suite.chargePoint.SendRequest(mockRequest)
	//TODO: currently the network error is not returned by SendRequest, but is only generated internally
	assert.Nil(t, err)
	// Assert that pending request was removed
	time.Sleep(500 * time.Millisecond)
	_, ok := suite.chargePoint.GetPendingRequest(callID)
	assert.False(t, ok)
}

// ----------------- SendResponse tests -----------------

func (suite *OcppJTestSuite) TestChargePointSendConfirmation() {
	t := suite.T()
	mockUniqueId := "1234"
	suite.mockClient.On("Write", mock.Anything).Return(nil)
	suite.mockClient.On("Start", mock.AnythingOfType("string")).Return(nil)
	_ = suite.chargePoint.Start("someUrl")
	mockConfirmation := newMockConfirmation("mockValue")
	// This is allowed. Endpoint doesn't keep track of incoming requests, but only outgoing ones
	err := suite.chargePoint.SendResponse(mockUniqueId, mockConfirmation)
	assert.Nil(t, err)
}

func (suite *OcppJTestSuite) TestChargePointSendInvalidConfirmation() {
	t := suite.T()
	mockUniqueId := "6789"
	suite.mockClient.On("Write", mock.Anything).Return(nil)
	suite.mockClient.On("Start", mock.AnythingOfType("string")).Return(nil)
	_ = suite.chargePoint.Start("someUrl")
	mockConfirmation := newMockConfirmation("")
	// This is allowed. Endpoint doesn't keep track of incoming requests, but only outgoing ones
	err := suite.chargePoint.SendResponse(mockUniqueId, mockConfirmation)
	assert.NotNil(t, err)
}

func (suite *OcppJTestSuite) TestChargePointSendConfirmationFailed() {
	t := suite.T()
	mockUniqueId := "1234"
	suite.mockClient.On("Write", mock.Anything).Return(errors.New("networkError"))
	suite.mockClient.On("Start", mock.AnythingOfType("string")).Return(nil)
	_ = suite.chargePoint.Start("someUrl")
	mockConfirmation := newMockConfirmation("mockValue")
	err := suite.chargePoint.SendResponse(mockUniqueId, mockConfirmation)
	assert.NotNil(t, err)
	assert.Equal(t, "networkError", err.Error())
}

// ----------------- SendError tests -----------------

func (suite *OcppJTestSuite) TestChargePointSendError() {
	t := suite.T()
	mockUniqueId := "1234"
	mockDescription := "mockDescription"
	suite.mockClient.On("Write", mock.Anything).Return(nil)
	err := suite.chargePoint.SendError(mockUniqueId, ocppj.GenericError, mockDescription, nil)
	assert.Nil(t, err)
}

func (suite *OcppJTestSuite) TestChargePointSendInvalidError() {
	t := suite.T()
	mockUniqueId := "6789"
	mockDescription := "mockDescription"
	suite.mockClient.On("Write", mock.Anything).Return(nil)
	err := suite.chargePoint.SendError(mockUniqueId, "InvalidErrorCode", mockDescription, nil)
	assert.NotNil(t, err)
}

func (suite *OcppJTestSuite) TestChargePointSendErrorFailed() {
	t := suite.T()
	mockUniqueId := "1234"
	suite.mockClient.On("Write", mock.Anything).Return(errors.New("networkError"))
	mockConfirmation := newMockConfirmation("mockValue")
	err := suite.chargePoint.SendResponse(mockUniqueId, mockConfirmation)
	assert.NotNil(t, err)
	assert.Equal(t, "networkError", err.Error())
}

// ----------------- Call Handlers tests -----------------

func (suite *OcppJTestSuite) TestChargePointCallHandler() {
	t := suite.T()
	mockUniqueId := "5678"
	mockValue := "someValue"
	mockRequest := fmt.Sprintf(`[2,"%v","%v",{"mockValue":"%v"}]`, mockUniqueId, MockFeatureName, mockValue)
	suite.chargePoint.SetRequestHandler(func(request ocpp.Request, requestId string, action string) {
		assert.Equal(t, mockUniqueId, requestId)
		assert.Equal(t, MockFeatureName, action)
		assert.NotNil(t, request)
	})
	suite.mockClient.On("Start", mock.AnythingOfType("string")).Return(nil).Run(func(args mock.Arguments) {
		// Simulate central system message
		err := suite.mockClient.MessageHandler([]byte(mockRequest))
		assert.Nil(t, err)
	})
	err := suite.chargePoint.Start("somePath")
	assert.Nil(t, err)
}

func (suite *OcppJTestSuite) TestChargePointCallResultHandler() {
	t := suite.T()
	mockUniqueId := "5678"
	mockValue := "someValue"
	mockRequest := newMockRequest("testValue")
	mockConfirmation := fmt.Sprintf(`[3,"%v",{"mockValue":"%v"}]`, mockUniqueId, mockValue)
	suite.chargePoint.SetResponseHandler(func(confirmation ocpp.Response, requestId string) {
		assert.Equal(t, mockUniqueId, requestId)
		assert.NotNil(t, confirmation)
	})
	suite.mockClient.On("Start", mock.AnythingOfType("string")).Return(nil)
	suite.chargePoint.AddPendingRequest(mockUniqueId, mockRequest) // Manually add a pending request, so that response is not rejected
	err := suite.chargePoint.Start("somePath")
	assert.Nil(t, err)
	// Simulate central system message
	err = suite.mockClient.MessageHandler([]byte(mockConfirmation))
	assert.Nil(t, err)
}

func (suite *OcppJTestSuite) TestChargePointCallErrorHandler() {
	t := suite.T()
	mockUniqueId := "5678"
	mockErrorCode := ocppj.GenericError
	mockErrorDescription := "Mock Description"
	mockValue := "someValue"
	mockErrorDetails := make(map[string]interface{})
	mockErrorDetails["details"] = "someValue"

	mockRequest := newMockRequest("testValue")
	mockError := fmt.Sprintf(`[4,"%v","%v","%v",{"details":"%v"}]`, mockUniqueId, mockErrorCode, mockErrorDescription, mockValue)
	suite.chargePoint.SetErrorHandler(func(err *ocpp.Error, details interface{}) {
		assert.Equal(t, mockUniqueId, err.MessageId)
		assert.Equal(t, mockErrorCode, err.Code)
		assert.Equal(t, mockErrorDescription, err.Description)
		assert.Equal(t, mockErrorDetails, details)
	})
	suite.mockClient.On("Start", mock.AnythingOfType("string")).Return(nil)
	suite.chargePoint.AddPendingRequest(mockUniqueId, mockRequest) // Manually add a pending request, so that response is not rejected
	err := suite.chargePoint.Start("someUrl")
	assert.Nil(t, err)
	// Simulate central system message
	err = suite.mockClient.MessageHandler([]byte(mockError))
	assert.Nil(t, err)
}

// ----------------- Queue processing tests -----------------

func (suite *OcppJTestSuite) TestClientEnqueueRequest() {
	t := suite.T()
	suite.mockClient.On("Start", mock.AnythingOfType("string")).Return(nil)
	suite.mockClient.On("Write", mock.Anything).Return(nil)
	// Start normally
	err := suite.chargePoint.Start("someUrl")
	require.Nil(t, err)
	req := newMockRequest("somevalue")
	err = suite.chargePoint.SendRequest(req)
	require.Nil(t, err)
	time.Sleep(500 * time.Millisecond)
	// Message was sent, but element should still be in queue
	require.False(t, suite.clientRequestQueue.IsEmpty())
	assert.Equal(t, 1, suite.clientRequestQueue.Size())
	// Analyze enqueued bundle
	peeked := suite.clientRequestQueue.Peek()
	require.NotNil(t, peeked)
	bundle, ok := peeked.(ocppj.RequestBundle)
	require.True(t, ok)
	require.NotNil(t, bundle)
	assert.Equal(t, req.GetFeatureName(), bundle.Call.Action)
	marshaled, err := bundle.Call.MarshalJSON()
	require.Nil(t, err)
	assert.Equal(t, marshaled, bundle.Data)
}

func (suite *OcppJTestSuite) TestClientEnqueueMultipleRequests() {
	t := suite.T()
	messagesToQueue := 5
	sentMessages := 0
	suite.mockClient.On("Start", mock.AnythingOfType("string")).Return(nil)
	suite.mockClient.On("Write", mock.Anything).Run(func(args mock.Arguments) {
		sentMessages += 1
	}).Return(nil)
	// Start normally
	err := suite.chargePoint.Start("someUrl")
	require.Nil(t, err)
	for i := 0; i < messagesToQueue; i++ {
		req := newMockRequest(fmt.Sprintf("request-%v", i))
		err = suite.chargePoint.SendRequest(req)
		require.Nil(t, err)
	}
	time.Sleep(500 * time.Millisecond)
	// Only one message was sent, but all elements should still be in queue
	assert.Equal(t, 1, sentMessages)
	require.False(t, suite.clientRequestQueue.IsEmpty())
	assert.Equal(t, messagesToQueue, suite.clientRequestQueue.Size())
	// Analyze enqueued bundle
	var i = 0
	for !suite.clientRequestQueue.IsEmpty() {
		popped := suite.clientRequestQueue.Pop()
		require.NotNil(t, popped)
		bundle, ok := popped.(ocppj.RequestBundle)
		require.True(t, ok)
		require.NotNil(t, bundle)
		assert.Equal(t, MockFeatureName, bundle.Call.Action)
		i++
	}
	assert.Equal(t, messagesToQueue, i)
}

func (suite *OcppJTestSuite) TestClientRequestQueueFull() {
	t := suite.T()
	messagesToQueue := queueCapacity
	suite.mockClient.On("Start", mock.AnythingOfType("string")).Return(nil)
	suite.mockClient.On("Write", mock.Anything).Return(nil)
	// Start normally
	err := suite.chargePoint.Start("someUrl")
	require.Nil(t, err)
	for i := 0; i < messagesToQueue; i++ {
		req := newMockRequest(fmt.Sprintf("request-%v", i))
		err = suite.chargePoint.SendRequest(req)
		require.Nil(t, err)
	}
	// Queue is now full. Trying to send an additional message should throw an error
	req := newMockRequest("full")
	err = suite.chargePoint.SendRequest(req)
	require.NotNil(t, err)
	assert.Equal(t, "request queue is full, cannot push new element", err.Error())
}

func (suite *OcppJTestSuite) TestClientParallelRequests() {
	t := suite.T()
	messagesToQueue := 10
	sentMessages := 0
	suite.mockClient.On("Start", mock.AnythingOfType("string")).Return(nil)
	suite.mockClient.On("Write", mock.Anything).Run(func(args mock.Arguments) {
		sentMessages += 1
	}).Return(nil)
	// Start normally
	err := suite.chargePoint.Start("someUrl")
	require.Nil(t, err)
	for i := 0; i < messagesToQueue; i++ {
		go func() {
			req := newMockRequest(fmt.Sprintf("someReq"))
			err = suite.chargePoint.SendRequest(req)
			require.Nil(t, err)
		}()
	}
	time.Sleep(1000 * time.Millisecond)
	// Only one message was sent, but all element should still be in queue
	require.False(t, suite.clientRequestQueue.IsEmpty())
	assert.Equal(t, messagesToQueue, suite.clientRequestQueue.Size())
}

// TestClientRequestFlow tests a typical flow with multiple request-responses.
//
// Requests are sent concurrently and a response to each message is sent from the mocked server endpoint.
// Both CallResult and CallError messages are returned to test all message types.
func (suite *OcppJTestSuite) TestClientRequestFlow() {
	t := suite.T()
	var mutex sync.Mutex
	messagesToQueue := 10
	processedMessages := 0
	sendResponseTrigger := make(chan *ocppj.Call, 1)
	suite.mockClient.On("Start", mock.AnythingOfType("string")).Return(nil)
	suite.mockClient.On("Write", mock.Anything).Run(func(args mock.Arguments) {
		data := args.Get(0).([]byte)
		call := ParseCall(&suite.chargePoint.Endpoint, string(data), t)
		require.NotNil(t, call)
		sendResponseTrigger <- call
	}).Return(nil)
	// Mocked response generator
	var wg sync.WaitGroup
	wg.Add(messagesToQueue)
	go func() {
		for {
			call, ok := <-sendResponseTrigger
			if !ok {
				// Test completed, quitting
				return
			}
			// Get original request to generate meaningful response
			peeked := suite.clientRequestQueue.Peek()
			bundle, _ := peeked.(ocppj.RequestBundle)
			require.NotNil(t, bundle)
			assert.Equal(t, call.UniqueId, bundle.Call.UniqueId)
			req, _ := call.Payload.(*MockRequest)
			// Send response back to client
			var data []byte
			var err error
			v, _ := strconv.Atoi(req.MockValue)
			if v%2 == 0 {
				// Send CallResult
				resp := newMockConfirmation("someResp")
				res, err := suite.chargePoint.CreateCallResult(resp, call.GetUniqueId())
				require.Nil(t, err)
				data, err = res.MarshalJSON()
				require.Nil(t, err)
			} else {
				// Send CallError
				res := suite.chargePoint.CreateCallError(call.GetUniqueId(), ocppj.GenericError, fmt.Sprintf("error-%v", req.MockValue), nil)
				data, err = res.MarshalJSON()
				require.Nil(t, err)
			}
			fmt.Printf("sending mocked response to message %v\n", call.GetUniqueId())
			err = suite.mockClient.MessageHandler(data) // Triggers ocppMessageHandler
			require.Nil(t, err)
			// Make sure the top queue element was popped
			mutex.Lock()
			processedMessages += 1
			peeked = suite.clientRequestQueue.Peek()
			if peeked != nil {
				bundle, _ := peeked.(ocppj.RequestBundle)
				require.NotNil(t, bundle)
				assert.NotEqual(t, call.UniqueId, bundle.Call.UniqueId)
			}
			mutex.Unlock()
			wg.Done()
		}
	}()
	// Start client normally
	err := suite.chargePoint.Start("someUrl")
	require.Nil(t, err)
	for i := 0; i < messagesToQueue; i++ {
		go func(j int) {
			req := newMockRequest(fmt.Sprintf("%v", j))
			err = suite.chargePoint.SendRequest(req)
			require.Nil(t, err)
		}(i)
	}
	// Wait for processing to complete
	wg.Wait()
	close(sendResponseTrigger)
	assert.True(t, suite.clientRequestQueue.IsEmpty())
}

//TODO: test retransmission
